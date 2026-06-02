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
	"time"

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
	// Messages API requires max_tokens to be set.
	defaultMaxTokens = 4096
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
// [provider.TooManyRequests].
func mapErr(err error) error {
	if err == nil {
		return nil
	}

	var statusErr xhttp.UnexpectedStatusCodeError
	if errors.As(err, &statusErr) {
		switch statusErr.StatusCode {
		case http.StatusTooManyRequests:
			return provider.TooManyRequests
		}
	}

	return err
}

// ----- wire protocol types -----

type apiRequest struct {
	Model         string            `json:"model"`
	MaxTokens     int               `json:"max_tokens"`
	System        string            `json:"system,omitempty"`
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
type apiContent struct {
	Type string `json:"type"`

	// type == "text"
	Text string `json:"text,omitempty"`

	// type == "thinking"
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`

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
}

type apiSource struct {
	Type      string `json:"type"` // "base64" | "url" | "file"
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
	FileID    string `json:"file_id,omitempty"`
}

type apiTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
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

