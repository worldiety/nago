// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/pkg/xhttp"
)

// DefaultBaseURL is the public OpenAI API endpoint, used whenever [Settings.BaseURL] is empty.
const DefaultBaseURL = "https://api.openai.com/v1"

// Client is a minimal, dependency-free client for the OpenAI Chat Completions API implemented directly on top
// of [xhttp]. It intentionally does not use the official openai-go SDK, to avoid pulling in unaudited
// transitive dependencies, and to stay tolerant against the many OpenAI-compatible servers which implement only
// a subset of the API.
type Client struct {
	c       *http.Client
	token   string
	base    string
	retry   int
	timeout time.Duration
	group   *xhttp.RequestGroup
}

// NewClient creates a new client. An empty baseURL selects [DefaultBaseURL], an empty token sends no
// Authorization header at all (as expected by local servers). rps <= 0 disables the client side rate limiter.
func NewClient(baseURL, token string, rps int, debug bool) *Client {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	timeout := 10 * time.Minute // large model responses, and local models on a CPU, can take a while

	return &Client{
		c:       &http.Client{Timeout: timeout},
		base:    strings.TrimRight(baseURL, "/") + "/",
		token:   strings.TrimSpace(token),
		retry:   2,
		timeout: timeout,
		group:   xhttp.NewRequestGroup().DebugLog(debug).RateLimit(rps),
	}
}

// isOpenAI reports whether the client talks to the official OpenAI API, which is stricter than most
// compatible servers (e.g. it rejects max_tokens for reasoning models).
func (c *Client) isOpenAI() bool {
	u, err := url.Parse(c.base)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Hostname(), "api.openai.com")
}

func (c *Client) newReq() *xhttp.Request {
	r := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		Timeout(c.timeout).
		Group(c.group)

	if c.token != "" {
		r = r.BearerAuthentication(c.token)
	}

	return r
}

// ChatCompletion performs a blocking POST /chat/completions call. Cancelling ctx aborts the request.
func (c *Client) ChatCompletion(ctx context.Context, req apiRequest) (apiResponse, error) {
	req.Stream = false
	req.StreamOptions = nil

	var resp apiResponse
	err := c.newReq().
		Context(ctx).
		URL("chat/completions").
		Assert2xx(true).
		BodyJSON(req).
		ToJSON(&resp).
		ToLimit(16 * 1024 * 1024).
		Post()
	if err != nil {
		return resp, mapErr(err)
	}

	// Some compatible servers report failures with a 200 status and an error object.
	if resp.Error != nil {
		return resp, resp.Error.asError(http.StatusOK)
	}

	return resp, nil
}

// ChatCompletionStream performs a streaming POST /chat/completions call. onChunk is invoked synchronously for
// every decoded chunk. Cancelling ctx aborts the request, including a response which is still streaming.
func (c *Client) ChatCompletionStream(ctx context.Context, req apiRequest, onChunk func(chunk apiChunk) error) error {
	req.Stream = true

	var cbErr error
	err := c.newReq().
		Context(ctx).
		URL("chat/completions").
		Assert2xx(true).
		BodyJSON(req).
		Header("Accept", "text/event-stream").
		ToCloser(func(rc io.ReadCloser) {
			defer func() { _ = rc.Close() }()
			cbErr = parseSSE(rc, func(data []byte) error {
				var chunk apiChunk
				if err := json.Unmarshal(data, &chunk); err != nil {
					return fmt.Errorf("decode stream chunk %q: %w", truncate(string(data), 256), err)
				}
				if chunk.Error != nil {
					return chunk.Error.asError(http.StatusOK)
				}
				return onChunk(chunk)
			})
		}).
		Post()
	if err != nil {
		return mapErr(err)
	}

	if cbErr != nil && ctx.Err() != nil {
		// A cancelled context surfaces as a read error of the body; report the cause instead.
		return fmt.Errorf("%w: %w", ctx.Err(), cbErr)
	}

	return cbErr
}

// ListModels returns the models from GET /models.
func (c *Client) ListModels() ([]apiModel, error) {
	var resp struct {
		Data []apiModel `json:"data"`
	}

	err := c.newReq().
		URL("models").
		Assert2xx(true).
		ToJSON(&resp).
		ToLimit(4 * 1024 * 1024).
		Get()

	return resp.Data, mapErr(err)
}

// ----- errors -----

// APIError is an error response of the API, e.g. {"error":{"message":"...","type":"...","code":"..."}}. It is
// returned for every non-2xx status which has no dedicated mapping (see [mapErr]) and unwraps to the underlying
// [xhttp.UnexpectedStatusCodeError], if any.
type APIError struct {
	StatusCode int
	Message    string
	Type       string
	Code       string
	cause      error
}

func (e *APIError) Error() string {
	var sb strings.Builder
	sb.WriteString("openai api error")
	if e.StatusCode != 0 {
		sb.WriteString(" (status " + strconv.Itoa(e.StatusCode))
		if e.Code != "" {
			sb.WriteString(", " + e.Code)
		} else if e.Type != "" {
			sb.WriteString(", " + e.Type)
		}
		sb.WriteString(")")
	}
	if e.Message != "" {
		sb.WriteString(": " + e.Message)
	}
	return sb.String()
}

func (e *APIError) Unwrap() error {
	return e.cause
}

// apiError is the wire representation of an error object. code is a string at OpenAI, but a number at some
// compatible servers (e.g. llama.cpp server), hence the raw message.
type apiError struct {
	Message string          `json:"message"`
	Type    string          `json:"type"`
	Code    json.RawMessage `json:"code"`
	// llama.cpp server reports the numbers of a context overflow as extra fields.
	NPromptTokens int `json:"n_prompt_tokens"`
	NCtx          int `json:"n_ctx"`
}

func (e *apiError) code() string {
	s := strings.TrimSpace(string(e.Code))
	if s == "" || s == "null" {
		return ""
	}
	var str string
	if err := json.Unmarshal(e.Code, &str); err == nil {
		return str
	}
	return s
}

func (e *apiError) asError(status int) error {
	if cwe, ok := e.contextWindowError(); ok {
		return cwe
	}
	return &APIError{StatusCode: status, Message: e.Message, Type: e.Type, Code: e.code()}
}

var (
	// maxContextRe matches the limit in the overflow messages of OpenAI, vLLM and OpenRouter, e.g. "This
	// model's maximum context length is 8192 tokens".
	maxContextRe = regexp.MustCompile(`(?i)maximum context length is (\d+) tokens`)
	// requestedTokensRe matches the size of the rejected request, e.g. "However, you requested 9000 tokens",
	// "your messages resulted in 9000 tokens" or "you requested about 9000 tokens".
	requestedTokensRe = regexp.MustCompile(`(?i)(?:requested|resulted in|requested about) (\d+) tokens`)
	// contextMarkerRe recognizes a context window overflow in the message, independent of the server.
	contextMarkerRe = regexp.MustCompile(`(?i)context[_ ]length[_ ]exceeded|maximum context length|exceeds? the (?:available )?context|context (?:size|window|length) (?:has been |was )?exceeded|prompt is too long|too many tokens|exceed_context_size`)
)

// contextWindowError detects a context window overflow and extracts the reported numbers when present.
func (e *apiError) contextWindowError() (completion.ContextWindowError, bool) {
	code := e.code()
	text := e.Message + " " + e.Type + " " + code
	if code != "context_length_exceeded" && !contextMarkerRe.MatchString(text) {
		return completion.ContextWindowError{}, false
	}

	cwe := completion.ContextWindowError{Limit: e.NCtx, Tokens: e.NPromptTokens}
	if m := maxContextRe.FindStringSubmatch(e.Message); m != nil && cwe.Limit == 0 {
		cwe.Limit, _ = strconv.Atoi(m[1])
	}
	if m := requestedTokensRe.FindStringSubmatch(e.Message); m != nil && cwe.Tokens == 0 {
		cwe.Tokens, _ = strconv.Atoi(m[1])
	}

	return cwe, true
}

// mapErr translates transport errors into the provider's defined error set: HTTP 429 becomes
// [completion.TooManyRequests], a recognizable context overflow becomes a [completion.ContextWindowError] and
// every other status an [APIError] carrying the message of the response body.
func mapErr(err error) error {
	if err == nil {
		return nil
	}

	var statusErr xhttp.UnexpectedStatusCodeError
	if !errors.As(err, &statusErr) {
		return err
	}

	e := parseErrorBody(statusErr.Body)

	if statusErr.StatusCode == http.StatusTooManyRequests {
		if e.Message != "" {
			return fmt.Errorf("%w: %s", completion.TooManyRequests, e.Message)
		}
		return completion.TooManyRequests
	}

	if statusErr.StatusCode >= 400 && statusErr.StatusCode < 500 {
		if cwe, ok := e.contextWindowError(); ok {
			return cwe
		}
	}

	return &APIError{
		StatusCode: statusErr.StatusCode,
		Message:    e.Message,
		Type:       e.Type,
		Code:       e.code(),
		cause:      statusErr,
	}
}

// parseErrorBody extracts the error object of a response body. It understands the OpenAI shape
// {"error":{...}}, the flat shape {"error":"message"} used by some compatible servers (e.g. Ollama) and falls
// back to the raw body as message.
func parseErrorBody(body []byte) apiError {
	var nested struct {
		Error json.RawMessage `json:"error"`
		// FastAPI based servers (e.g. vLLM on validation errors) use "detail" or a top-level "message".
		Message string `json:"message"`
		Detail  any    `json:"detail"`
	}

	if err := json.Unmarshal(body, &nested); err == nil {
		var obj apiError
		if len(nested.Error) > 0 && nested.Error[0] == '{' {
			if err := json.Unmarshal(nested.Error, &obj); err == nil && obj.Message != "" {
				return obj
			}
		}

		var str string
		if len(nested.Error) > 0 && json.Unmarshal(nested.Error, &str) == nil && str != "" {
			return apiError{Message: str}
		}

		if nested.Message != "" {
			return apiError{Message: nested.Message}
		}

		if nested.Detail != nil {
			if s, ok := nested.Detail.(string); ok {
				return apiError{Message: s}
			}
			b, _ := json.Marshal(nested.Detail)
			return apiError{Message: string(b)}
		}
	}

	return apiError{Message: truncate(string(bytes.TrimSpace(body)), 1024)}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// ----- wire protocol types -----

type apiRequest struct {
	Model               string            `json:"model"`
	Messages            []apiMessage      `json:"messages"`
	Tools               []apiTool         `json:"tools,omitempty"`
	ToolChoice          any               `json:"tool_choice,omitempty"`
	MaxTokens           int               `json:"max_tokens,omitempty"`
	MaxCompletionTokens int               `json:"max_completion_tokens,omitempty"`
	Temperature         *float64          `json:"temperature,omitempty"`
	TopP                *float64          `json:"top_p,omitempty"`
	Stop                []string          `json:"stop,omitempty"`
	Stream              bool              `json:"stream,omitempty"`
	StreamOptions       *apiStreamOptions `json:"stream_options,omitempty"`
}

type apiStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// apiMessage is a single chat message. Content is either a plain string, a list of [apiPart]s or nil (for an
// assistant message which only carries tool calls), therefore it is typed as any.
type apiMessage struct {
	Role       string        `json:"role"`
	Content    any           `json:"content"`
	ToolCalls  []apiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

// apiPart is a content part of a user message.
type apiPart struct {
	Type     string       `json:"type"` // "text" | "image_url" | "file"
	Text     string       `json:"text,omitempty"`
	ImageURL *apiImageURL `json:"image_url,omitempty"`
	File     *apiFile     `json:"file,omitempty"`
}

type apiImageURL struct {
	URL string `json:"url"`
}

type apiFile struct {
	FileName string `json:"filename,omitempty"`
	FileData string `json:"file_data,omitempty"`
	FileID   string `json:"file_id,omitempty"`
}

type apiTool struct {
	Type     string      `json:"type"` // always "function"
	Function apiFunction `json:"function"`
}

type apiFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type apiToolCall struct {
	// Index identifies the call within a streamed response, where a call is spread over several chunks.
	Index    *int                `json:"index,omitempty"`
	ID       string              `json:"id,omitempty"`
	Type     string              `json:"type,omitempty"`
	Function apiToolCallFunction `json:"function"`
}

type apiToolCallFunction struct {
	Name string `json:"name,omitempty"`
	// Arguments is a JSON document encoded as string. Some compatible servers send a JSON object instead,
	// see [apiToolCallFunction.UnmarshalJSON].
	Arguments string `json:"arguments"`
}

// UnmarshalJSON accepts the arguments either as the specified JSON string or as a raw JSON object, which some
// compatible servers emit.
func (f *apiToolCallFunction) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	f.Name = raw.Name
	f.Arguments = ""

	args := bytes.TrimSpace(raw.Arguments)
	switch {
	case len(args) == 0 || string(args) == "null":
	case args[0] == '"':
		if err := json.Unmarshal(args, &f.Arguments); err != nil {
			return err
		}
	default:
		f.Arguments = string(args)
	}

	return nil
}

type apiResponse struct {
	ID      string      `json:"id"`
	Model   string      `json:"model"`
	Choices []apiChoice `json:"choices"`
	Usage   *apiUsage   `json:"usage"`
	Error   *apiError   `json:"error"`
}

type apiChoice struct {
	Index        int                `json:"index"`
	Message      apiResponseMessage `json:"message"`
	FinishReason string             `json:"finish_reason"`
}

// apiResponseMessage is the assistant message of a response and, in a streamed response, the delta of a
// chunk. Reasoning is reported by compatible servers in different, non-standard fields: reasoning_content
// (DeepSeek, vLLM, llama.cpp server) and reasoning (Ollama, OpenRouter).
type apiResponseMessage struct {
	Role             string        `json:"role"`
	Content          *string       `json:"content"`
	Refusal          *string       `json:"refusal"`
	ReasoningContent *string       `json:"reasoning_content"`
	Reasoning        *string       `json:"reasoning"`
	ToolCalls        []apiToolCall `json:"tool_calls"`
}

func (m apiResponseMessage) reasoning() string {
	if m.ReasoningContent != nil && *m.ReasoningContent != "" {
		return *m.ReasoningContent
	}
	if m.Reasoning != nil {
		return *m.Reasoning
	}
	return ""
}

type apiChunk struct {
	ID      string           `json:"id"`
	Model   string           `json:"model"`
	Choices []apiChunkChoice `json:"choices"`
	Usage   *apiUsage        `json:"usage"`
	Error   *apiError        `json:"error"`
}

type apiChunkChoice struct {
	Index        int                `json:"index"`
	Delta        apiResponseMessage `json:"delta"`
	FinishReason *string            `json:"finish_reason"`
}

type apiUsage struct {
	PromptTokens        int `json:"prompt_tokens"`
	CompletionTokens    int `json:"completion_tokens"`
	PromptTokensDetails *struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
}

type apiModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}
