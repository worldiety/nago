// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/pkg/xhttp"
)

const (
	// defaultBaseURL is the public Anthropic API endpoint.
	defaultBaseURL = "https://api.anthropic.com/v1/"
	// defaultVersion is the value of the required anthropic-version header. See
	// https://docs.anthropic.com/en/api/versioning.
	defaultVersion = "2023-06-01"
	// defaultMaxTokens is used whenever neither the request nor the settings provide a value. The Anthropic
	// Messages API requires max_tokens to be set. We default fairly high so agentic turns with many
	// (parallel) tool_use blocks are not truncated (stop_reason == max_tokens), which would otherwise leave
	// a tool_use without a matching tool_result. Older models (claude-3-opus/haiku) cap output at 4096;
	// for those set Settings.MaxTokens explicitly.
	defaultMaxTokens = 8192
)

// Client is a minimal, dependency-free Anthropic API client implemented directly on top of [xhttp].
// It intentionally does not use the official anthropic-sdk-go to avoid pulling in unaudited transitive
// dependencies.
type Client struct {
	c       *http.Client
	token   string
	base    string
	version string
	retry   int
	debug   bool
	timeout time.Duration
	group   *xhttp.RequestGroup
}

// NewClient creates a new Anthropic client. version may be empty to use [defaultVersion]. rps <= 0 disables
// the client side rate limiter.
func NewClient(token, version string, rps int, debug bool) *Client {
	if version == "" {
		version = defaultVersion
	}

	timeout := 10 * time.Minute // large model responses can take a while

	return &Client{
		c:       &http.Client{Timeout: timeout},
		base:    defaultBaseURL,
		token:   token,
		version: version,
		retry:   2,
		debug:   debug,
		timeout: timeout,
		group:   xhttp.NewRequestGroup().DebugLog(debug).RateLimit(rps),
	}
}

func (c *Client) newReq() *xhttp.Request {
	return xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		Timeout(c.timeout).
		Group(c.group).
		Header("x-api-key", c.token).
		Header("anthropic-version", c.version)
}

// mapErr translates transport errors into the provider's defined error set, e.g. HTTP 429 to
// [provider.TooManyRequests] and the HTTP 400 "prompt is too long" overflow to
// [completion.ContextWindowError].
func mapErr(err error) error {
	if err == nil {
		return nil
	}

	var statusErr xhttp.UnexpectedStatusCodeError
	if errors.As(err, &statusErr) {
		switch statusErr.StatusCode {
		case http.StatusTooManyRequests:
			return provider.TooManyRequests
		case http.StatusBadRequest:
			if cwe, ok := parseContextWindowError(statusErr.Body); ok {
				return cwe
			}
		}
	}

	return err
}

// promptTooLongRe matches the Anthropic overflow message, e.g.
// "prompt is too long: 215534 tokens > 200000 maximum". The numbers are optional from our perspective; only
// the "prompt is too long" marker is required to classify the error.
var promptTooLongRe = regexp.MustCompile(`prompt is too long(?:: (\d+) tokens > (\d+) maximum)?`)

// parseContextWindowError detects Anthropic's context window overflow in a 400 response body and extracts the
// reported token count and limit when present.
func parseContextWindowError(body []byte) (completion.ContextWindowError, bool) {
	m := promptTooLongRe.FindSubmatch(body)
	if m == nil {
		return completion.ContextWindowError{}, false
	}

	var cwe completion.ContextWindowError
	if len(m[1]) > 0 {
		cwe.Tokens, _ = strconv.Atoi(string(m[1]))
	}
	if len(m[2]) > 0 {
		cwe.Limit, _ = strconv.Atoi(string(m[2]))
	}

	return cwe, true
}

// ----- wire protocol types -----

// apiCacheControl marks a content block as a prompt-cache breakpoint. Everything in the request prefix up to
// and including the marked block becomes cacheable on Anthropic's servers. See
// https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching.
type apiCacheControl struct {
	Type string `json:"type"`          // always "ephemeral"
	TTL  string `json:"ttl,omitempty"` // "5m" (default) | "1h"
}

type apiRequest struct {
	Model         string            `json:"model"`
	MaxTokens     int               `json:"max_tokens"`
	System        []apiContent      `json:"system,omitempty"`
	Messages      []apiMessage      `json:"messages"`
	Tools         []apiTool         `json:"tools,omitempty"`
	ToolChoice    *apiToolChoice    `json:"tool_choice,omitempty"`
	Temperature   *float64          `json:"temperature,omitempty"`
	TopP          *float64          `json:"top_p,omitempty"`
	StopSequences []string          `json:"stop_sequences,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Stream        bool              `json:"stream,omitempty"`
}

type apiMessage struct {
	Role    string       `json:"role"`
	Content []apiContent `json:"content"`
}

// apiContent is the polymorphic content block used for both requests and responses. Only the fields
// relevant for the respective Type are populated.
//
// It carries a custom [apiContent.MarshalJSON] because the struct tags alone cannot express the
// Anthropic schema correctly: the API became stricter and requires the "thinking" (and "signature")
// fields to be present on every thinking block, even when the reasoning text is empty. Relying on
// "omitempty" silently dropped an empty "thinking" field and produced
// `messages[i].content[j].thinking.thinking: Field required`. The custom marshaller therefore emits
// exactly the fields that belong to the respective Type.
type apiContent struct {
	Type string `json:"type"`

	// type == "text"
	Text string `json:"text,omitempty"`

	// type == "thinking"
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`

	// type == "redacted_thinking"
	Data string `json:"data,omitempty"`

	// type == "image" | "document"
	Source *apiSource `json:"source,omitempty"`

	// type == "tool_use"
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// type == "tool_result"
	ToolUseID string       `json:"tool_use_id,omitempty"`
	Content   []apiContent `json:"content,omitempty"`
	IsError   bool         `json:"is_error,omitempty"`

	// CacheControl, when set, marks this block as a prompt-cache breakpoint.
	CacheControl *apiCacheControl `json:"cache_control,omitempty"`
}

// MarshalJSON renders the content block with only the fields relevant to its Type. This is required
// because Anthropic's Messages API validates each block against a per-type schema and now (a) demands
// that thinking blocks always carry their "thinking" and "signature" fields — even when the reasoning
// text is empty — and (b) rejects unrelated/empty sibling fields. A plain struct with "omitempty" tags
// cannot satisfy both constraints with a single shared struct, so we assemble the payload explicitly.
func (c apiContent) MarshalJSON() ([]byte, error) {
	m := map[string]any{"type": c.Type}

	switch c.Type {
	case "text":
		m["text"] = c.Text
	case "thinking":
		// thinking and signature are mandatory for thinking blocks; never drop them.
		m["thinking"] = c.Thinking
		m["signature"] = c.Signature
	case "redacted_thinking":
		m["data"] = c.Data
	case "image", "document":
		m["source"] = c.Source
	case "tool_use":
		m["id"] = c.ID
		m["name"] = c.Name
		input := c.Input
		if len(input) == 0 {
			input = json.RawMessage("{}")
		}
		m["input"] = input
	case "tool_result":
		m["tool_use_id"] = c.ToolUseID
		if c.Content != nil {
			m["content"] = c.Content
		}
		if c.IsError {
			m["is_error"] = true
		}
	default:
		// Unknown/forward-compatible block types: fall back to the non-empty known fields so we never
		// emit something the validator is guaranteed to reject.
		if c.Text != "" {
			m["text"] = c.Text
		}
	}

	if c.CacheControl != nil {
		m["cache_control"] = c.CacheControl
	}

	return json.Marshal(m)
}

type apiSource struct {
	Type      string `json:"type"` // "base64" | "url" | "file"
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
	FileID    string `json:"file_id,omitempty"`
}

type apiTool struct {
	Name         string           `json:"name"`
	Description  string           `json:"description,omitempty"`
	InputSchema  json.RawMessage  `json:"input_schema"`
	CacheControl *apiCacheControl `json:"cache_control,omitempty"`
}

type apiToolChoice struct {
	Type string `json:"type"` // "auto" | "any" | "tool" | "none"
	Name string `json:"name,omitempty"`
}

type apiResponse struct {
	ID         string       `json:"id"`
	Type       string       `json:"type"`
	Role       string       `json:"role"`
	Model      string       `json:"model"`
	Content    []apiContent `json:"content"`
	StopReason string       `json:"stop_reason"`
	Usage      apiUsage     `json:"usage"`
}

type apiUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

type apiModel struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
}

// CreateMessage performs a blocking, stateless Messages API call.
func (c *Client) CreateMessage(req apiRequest) (apiResponse, error) {
	req.Stream = false

	var resp apiResponse
	err := c.newReq().
		URL("messages").
		Assert2xx(true).
		BodyJSON(req).
		ToJSON(&resp).
		ToLimit(8 * 1024 * 1024).
		Post()

	return resp, mapErr(err)
}

// ListModels returns the models available for the configured token.
func (c *Client) ListModels() ([]apiModel, error) {
	var resp struct {
		Data []apiModel `json:"data"`
	}

	err := c.newReq().
		URL("models").
		Query("limit", "1000").
		Assert2xx(true).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp.Data, mapErr(err)
}

