// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package openai

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/pkg/xhttp"
	"go.wdy.de/nago/presentation/core"
)

// marshal renders v and decodes it again into a generic structure, so the tests assert the wire format.
func marshal(t *testing.T, v any) map[string]any {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestBuildRequest_Messages(t *testing.T) {
	p := newProvider("id", Settings{BaseURL: "http://localhost:11434/v1"})

	opts := completion.Options{
		Model:         "m",
		System:        "be nice",
		MaxTokens:     100,
		Temperature:   option.Some(0.5),
		TopP:          option.Some(0.9),
		StopSequences: []string{"END"},
		Tools: []completion.ToolDef{
			{Name: "weather", Description: "get the weather", Schema: json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`)},
			{Name: "noargs"},
		},
		ToolChoice: completion.ToolChoice{Mode: "any"},
		Messages: []completion.Message{
			{Role: completion.User, Content: []completion.Content{completion.Text{Text: "weather in Oldenburg?"}}},
			{Role: completion.Assistant, Content: []completion.Content{
				completion.Thinking{Text: "let me think", Signature: "sig"},
				completion.ToolCall{ID: "call_1", Name: "weather", Arguments: json.RawMessage(`{"city":"Oldenburg"}`)},
				completion.ToolCall{ID: "call_2", Name: "noargs"},
			}},
			{Role: completion.User, Content: []completion.Content{
				completion.ToolResult{ToolCallID: "call_1", Content: []completion.Content{completion.Text{Text: "sunny"}}},
				completion.ToolResult{ToolCallID: "call_2", IsError: true, Content: []completion.Content{completion.Text{Text: "boom"}}},
				completion.Text{Text: "and tomorrow?"},
			}},
		},
	}

	req, err := p.buildRequest(opts)
	if err != nil {
		t.Fatal(err)
	}

	m := marshal(t, req)
	if m["max_tokens"] != float64(100) || m["max_completion_tokens"] != nil {
		t.Errorf("expected max_tokens for a compatible server, got %v / %v", m["max_tokens"], m["max_completion_tokens"])
	}
	if m["temperature"] != 0.5 || m["top_p"] != 0.9 {
		t.Errorf("unexpected sampling params: %v %v", m["temperature"], m["top_p"])
	}
	if m["tool_choice"] != "required" {
		t.Errorf("expected tool_choice required, got %v", m["tool_choice"])
	}
	if stop := m["stop"].([]any); len(stop) != 1 || stop[0] != "END" {
		t.Errorf("unexpected stop: %v", m["stop"])
	}

	tools := m["tools"].([]any)
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	fn := tools[0].(map[string]any)["function"].(map[string]any)
	if tools[0].(map[string]any)["type"] != "function" || fn["name"] != "weather" || fn["parameters"].(map[string]any)["type"] != "object" {
		t.Errorf("unexpected tool: %v", tools[0])
	}
	if _, ok := tools[1].(map[string]any)["function"].(map[string]any)["parameters"]; ok {
		t.Errorf("expected parameters to be omitted for an empty schema")
	}

	msgs := m["messages"].([]any)
	roles := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		roles = append(roles, msg.(map[string]any)["role"].(string))
	}
	if got := strings.Join(roles, ","); got != "system,user,assistant,tool,tool,user" {
		t.Fatalf("unexpected role order: %s", got)
	}

	if msgs[0].(map[string]any)["content"] != "be nice" {
		t.Errorf("unexpected system message: %v", msgs[0])
	}
	if msgs[1].(map[string]any)["content"] != "weather in Oldenburg?" {
		t.Errorf("expected plain string user content, got %v", msgs[1])
	}

	asst := msgs[2].(map[string]any)
	if c, ok := asst["content"]; !ok || c != nil {
		t.Errorf("expected content null for a pure tool call turn (thinking dropped), got %v", asst["content"])
	}
	calls := asst["tool_calls"].([]any)
	call1 := calls[0].(map[string]any)
	if call1["id"] != "call_1" || call1["type"] != "function" || call1["function"].(map[string]any)["arguments"] != `{"city":"Oldenburg"}` {
		t.Errorf("unexpected tool call: %v", call1)
	}
	if args := calls[1].(map[string]any)["function"].(map[string]any)["arguments"]; args != "{}" {
		t.Errorf("expected empty arguments to become {}, got %v", args)
	}
	if _, ok := call1["index"]; ok {
		t.Errorf("index must not be sent in requests")
	}

	tool1 := msgs[3].(map[string]any)
	if tool1["tool_call_id"] != "call_1" || tool1["content"] != "sunny" {
		t.Errorf("unexpected tool message: %v", tool1)
	}
	if tool2 := msgs[4].(map[string]any); tool2["content"] != "Error: boom" {
		t.Errorf("expected error prefix, got %v", tool2)
	}
	if msgs[5].(map[string]any)["content"] != "and tomorrow?" {
		t.Errorf("unexpected trailing user message: %v", msgs[5])
	}
}

func TestBuildRequest_MaxTokensField(t *testing.T) {
	opts := completion.Options{Model: "m", Messages: []completion.Message{{Role: completion.User, Content: []completion.Content{completion.Text{Text: "hi"}}}}}

	openai := newProvider("id", Settings{})
	req, _ := openai.buildRequest(opts)
	if req.MaxTokens != 0 || req.MaxCompletionTokens != 0 {
		t.Errorf("expected no limit without settings, got %+v", req)
	}

	openai = newProvider("id", Settings{MaxTokens: 42})
	req, _ = openai.buildRequest(opts)
	if req.MaxCompletionTokens != 42 || req.MaxTokens != 0 {
		t.Errorf("expected max_completion_tokens for api.openai.com, got %+v", req)
	}

	opts.MaxTokens = 7
	local := newProvider("id", Settings{BaseURL: "http://localhost:8080/v1/", MaxTokens: 42})
	req, _ = local.buildRequest(opts)
	if req.MaxTokens != 7 || req.MaxCompletionTokens != 0 {
		t.Errorf("expected max_tokens=7 for a compatible server, got %+v", req)
	}
}

func TestBuildRequest_ToolChoice(t *testing.T) {
	cases := []struct {
		tc   completion.ToolChoice
		want string
	}{
		{completion.ToolChoice{}, `null`},
		{completion.ToolChoice{Mode: "auto"}, `"auto"`},
		{completion.ToolChoice{Mode: "none"}, `"none"`},
		{completion.ToolChoice{Mode: "any"}, `"required"`},
		{completion.ToolChoice{Name: "x"}, `{"function":{"name":"x"},"type":"function"}`},
	}

	for _, c := range cases {
		b, _ := json.Marshal(toAPIToolChoice(c.tc))
		if string(b) != c.want {
			t.Errorf("%+v: got %s, want %s", c.tc, b, c.want)
		}
	}

	// without tools, tool_choice must be omitted, OpenAI rejects it otherwise
	p := newProvider("id", Settings{})
	req, _ := p.buildRequest(completion.Options{Model: "m", ToolChoice: completion.ToolChoice{Mode: "any"}, Messages: []completion.Message{{Role: completion.User, Content: []completion.Content{completion.Text{Text: "hi"}}}}})
	if _, ok := marshal(t, req)["tool_choice"]; ok {
		t.Errorf("expected tool_choice to be omitted without tools")
	}
}

func TestBuildRequest_Media(t *testing.T) {
	p := newProvider("id", Settings{})
	opts := completion.Options{
		Model: "m",
		Messages: []completion.Message{
			{Role: completion.User, Content: []completion.Content{
				completion.Text{Text: "look"},
				completion.Media{MimeType: file.PNG, Source: completion.Source{Data: []byte("png")}},
				completion.Media{MimeType: file.JPEG, Source: completion.Source{URL: option.Some[core.URI]("https://example.com/a.jpg")}},
				completion.Media{MimeType: file.Markdown, Source: completion.Source{Data: []byte("# md")}},
				completion.Media{MimeType: file.PDF, Source: completion.Source{Data: []byte("%PDF")}},
				completion.Media{MimeType: file.DOCX, Source: completion.Source{FileID: option.Some[file.ID]("file_1")}},
			}},
			{Role: completion.Assistant, Content: []completion.Content{completion.ToolCall{ID: "c", Name: "shot"}}},
			{Role: completion.User, Content: []completion.Content{
				completion.ToolResult{ToolCallID: "c", Content: []completion.Content{
					completion.Text{Text: "here"},
					completion.Media{MimeType: file.GIF, Source: completion.Source{Data: []byte("gif")}},
				}},
			}},
		},
	}

	req, err := p.buildRequest(opts)
	if err != nil {
		t.Fatal(err)
	}

	msgs := marshal(t, req)["messages"].([]any)
	parts := msgs[0].(map[string]any)["content"].([]any)
	if len(parts) != 6 {
		t.Fatalf("expected 6 parts, got %d: %v", len(parts), parts)
	}

	part := func(i int) map[string]any { return parts[i].(map[string]any) }
	if part(1)["type"] != "image_url" || part(1)["image_url"].(map[string]any)["url"] != "data:image/png;base64,cG5n" {
		t.Errorf("unexpected inline image: %v", part(1))
	}
	if part(2)["image_url"].(map[string]any)["url"] != "https://example.com/a.jpg" {
		t.Errorf("unexpected url image: %v", part(2))
	}
	if part(3)["type"] != "text" || part(3)["text"] != "# md" {
		t.Errorf("expected textual media inline, got %v", part(3))
	}
	if part(4)["type"] != "file" || part(4)["file"].(map[string]any)["file_data"] != "data:application/pdf;base64,JVBERg==" {
		t.Errorf("unexpected pdf part: %v", part(4))
	}
	if part(5)["type"] != "text" || !strings.Contains(part(5)["text"].(string), "file_1") {
		t.Errorf("expected a placeholder for a foreign file id, got %v", part(5))
	}

	// the tool result image is moved into a user message following the tool message
	if msgs[2].(map[string]any)["role"] != "tool" || !strings.HasPrefix(msgs[2].(map[string]any)["content"].(string), "here\n") {
		t.Errorf("unexpected tool message: %v", msgs[2])
	}
	follow := msgs[3].(map[string]any)
	fparts := follow["content"].([]any)
	if follow["role"] != "user" || len(fparts) != 2 || fparts[1].(map[string]any)["type"] != "image_url" {
		t.Errorf("expected the tool image in a following user message, got %v", follow)
	}
}

func TestFromAPIResponse(t *testing.T) {
	body := `{
		"id":"x","model":"gpt-test",
		"choices":[{"index":0,"finish_reason":"tool_calls","message":{
			"role":"assistant","content":"calling","reasoning_content":"hmm",
			"tool_calls":[
				{"id":"call_a","type":"function","function":{"name":"weather","arguments":"{\"city\":\"Bremen\"}"}},
				{"type":"function","function":{"name":"obj","arguments":{"a":1}}}
			]}}],
		"usage":{"prompt_tokens":100,"completion_tokens":20,"prompt_tokens_details":{"cached_tokens":60}}
	}`

	var resp apiResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatal(err)
	}

	res, err := fromAPIResponse(resp, "requested")
	if err != nil {
		t.Fatal(err)
	}

	if res.StopReason != completion.StopToolUse || res.Model != "gpt-test" {
		t.Errorf("unexpected result: %+v", res)
	}
	if res.Usage != (completion.Usage{InputTokens: 40, OutputTokens: 20, CacheReadTokens: 60}) {
		t.Errorf("unexpected usage: %+v", res.Usage)
	}

	c := res.Message.Content
	if len(c) != 4 {
		t.Fatalf("expected 4 blocks, got %d: %+v", len(c), c)
	}
	if th, ok := c[0].(completion.Thinking); !ok || th.Text != "hmm" {
		t.Errorf("expected thinking first, got %+v", c[0])
	}
	if tx, ok := c[1].(completion.Text); !ok || tx.Text != "calling" {
		t.Errorf("expected text, got %+v", c[1])
	}
	if tc, ok := c[2].(completion.ToolCall); !ok || tc.ID != "call_a" || string(tc.Arguments) != `{"city":"Bremen"}` {
		t.Errorf("unexpected tool call: %+v", c[2])
	}
	if tc, ok := c[3].(completion.ToolCall); !ok || !strings.HasPrefix(tc.ID, "call_") || string(tc.Arguments) != `{"a":1}` {
		t.Errorf("expected a generated id and object arguments, got %+v", c[3])
	}
}

func TestFromAPIFinishReason(t *testing.T) {
	cases := []struct {
		reason string
		tools  bool
		want   completion.StopReason
	}{
		{"stop", false, completion.StopEndTurn},
		{"stop", true, completion.StopToolUse},
		{"", false, completion.StopEndTurn},
		{"length", true, completion.StopMaxTokens},
		{"tool_calls", false, completion.StopToolUse},
		{"function_call", false, completion.StopToolUse},
		{"content_filter", false, completion.StopRefusal},
		{"other", false, "other"},
	}
	for _, c := range cases {
		if got := fromAPIFinishReason(c.reason, c.tools); got != c.want {
			t.Errorf("%q/%v: got %q, want %q", c.reason, c.tools, got, c.want)
		}
	}
}

func TestMapErr(t *testing.T) {
	status := func(code int, body string) error {
		return xhttp.UnexpectedStatusCodeError{StatusCode: code, Body: []byte(body)}
	}

	if err := mapErr(status(429, `{"error":{"message":"slow down","type":"rate_limit"}}`)); !errors.Is(err, completion.TooManyRequests) {
		t.Errorf("expected TooManyRequests, got %v", err)
	}

	overflows := []struct {
		body          string
		limit, tokens int
	}{
		{`{"error":{"message":"This model's maximum context length is 8192 tokens. However, your messages resulted in 9000 tokens.","type":"invalid_request_error","code":"context_length_exceeded"}}`, 8192, 9000},
		{`{"object":"error","message":"This model's maximum context length is 2048 tokens. However, you requested 3000 tokens (2900 in the messages, 100 in the completion).","type":"BadRequestError","code":400}`, 2048, 3000},
		{`{"error":{"code":400,"message":"the request exceeds the available context size, try increasing it","type":"exceed_context_size_error","n_prompt_tokens":5000,"n_ctx":4096}}`, 4096, 5000},
		{`{"error":{"message":"Input is too long","code":"context_length_exceeded"}}`, 0, 0},
	}
	for _, o := range overflows {
		err := mapErr(status(400, o.body))
		var cwe completion.ContextWindowError
		if !errors.Is(err, completion.ContextWindowExceeded) || !errors.As(err, &cwe) {
			t.Errorf("expected context window error for %s, got %v", o.body, err)
			continue
		}
		if cwe.Limit != o.limit || cwe.Tokens != o.tokens {
			t.Errorf("unexpected numbers %+v for %s", cwe, o.body)
		}
	}

	err := mapErr(status(401, `{"error":{"message":"Incorrect API key provided","type":"invalid_request_error","code":"invalid_api_key"}}`))
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 401 || apiErr.Code != "invalid_api_key" || !strings.Contains(err.Error(), "Incorrect API key") {
		t.Errorf("unexpected error: %v", err)
	}
	var statusErr xhttp.UnexpectedStatusCodeError
	if !errors.As(err, &statusErr) {
		t.Errorf("expected the status error to be unwrappable")
	}

	if err := mapErr(status(404, `{"error":"model \"foo\" not found, try pulling it first"}`)); !strings.Contains(err.Error(), `model "foo" not found`) {
		t.Errorf("expected the flat ollama error message, got %v", err)
	}
	if err := mapErr(status(502, `Bad Gateway`)); !strings.Contains(err.Error(), "Bad Gateway") {
		t.Errorf("expected the raw body as message, got %v", err)
	}
}

func TestSettingsDecodeLegacy(t *testing.T) {
	// secrets stored by the former credentials-only package must still decode
	var s Settings
	if err := json.Unmarshal([]byte(`{"Name":"legacy","Token":"sk-123"}`), &s); err != nil {
		t.Fatal(err)
	}
	if s.Name != "legacy" || s.Token != "sk-123" || s.BaseURL != "" {
		t.Errorf("unexpected settings: %+v", s)
	}
}
