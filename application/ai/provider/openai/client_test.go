// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/completion"
)

// fakeServer is a scripted OpenAI-compatible server. handle receives the decoded request body.
type fakeServer struct {
	t      *testing.T
	srv    *httptest.Server
	mu     sync.Mutex
	bodies []map[string]any
	auth   []string
}

func newFakeServer(t *testing.T, handle func(w http.ResponseWriter, r *http.Request, body map[string]any)) *fakeServer {
	t.Helper()
	fs := &fakeServer{t: t}
	fs.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			if len(b) > 0 {
				if err := json.Unmarshal(b, &body); err != nil {
					t.Errorf("invalid request body: %v", err)
				}
			}
		}
		fs.mu.Lock()
		fs.bodies = append(fs.bodies, body)
		fs.auth = append(fs.auth, r.Header.Get("Authorization"))
		fs.mu.Unlock()
		handle(w, r, body)
	}))
	t.Cleanup(fs.srv.Close)
	return fs
}

func (fs *fakeServer) provider(token string) *openaiProvider {
	return newProvider("test", Settings{BaseURL: fs.srv.URL + "/v1", Token: token})
}

func (fs *fakeServer) body(i int) map[string]any {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.bodies[i]
}

func (fs *fakeServer) calls() int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return len(fs.bodies)
}

func writeJSON(w http.ResponseWriter, status int, v string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, v)
}

func writeSSE(w http.ResponseWriter, events ...string) {
	w.Header().Set("Content-Type", "text/event-stream")
	for _, e := range events {
		_, _ = fmt.Fprintf(w, "data: %s\n\n", e)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func userOpts(text string) completion.Options {
	return completion.Options{
		Model:    "test-model",
		Messages: []completion.Message{{Role: completion.User, Content: []completion.Content{completion.Text{Text: text}}}},
	}
}

func TestComplete(t *testing.T) {
	fs := newFakeServer(t, func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		if r.URL.Path != "/v1/chat/completions" || r.Method != http.MethodPost {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, `{"model":"test-model","choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"hello"}}],"usage":{"prompt_tokens":3,"completion_tokens":1}}`)
	})

	res, err := fs.provider("sk-secret").completions.Complete(context.Background(), nil, userOpts("hi"))
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Message.Content) != 1 || res.Message.Content[0].(completion.Text).Text != "hello" {
		t.Errorf("unexpected content: %+v", res.Message.Content)
	}
	if res.StopReason != completion.StopEndTurn || res.Usage.InputTokens != 3 || res.Usage.OutputTokens != 1 {
		t.Errorf("unexpected result: %+v", res)
	}
	if fs.auth[0] != "Bearer sk-secret" {
		t.Errorf("unexpected Authorization header %q", fs.auth[0])
	}
	if s, ok := fs.body(0)["stream"]; ok {
		t.Errorf("expected no stream flag, got %v", s)
	}

	// without a token, no Authorization header is sent at all
	if _, err := fs.provider("").completions.Complete(context.Background(), nil, userOpts("hi")); err != nil {
		t.Fatal(err)
	}
	if fs.auth[1] != "" {
		t.Errorf("expected no Authorization header, got %q", fs.auth[1])
	}
}

func TestComplete_Errors(t *testing.T) {
	cases := []struct {
		status int
		body   string
		check  func(error) bool
	}{
		{429, `{"error":{"message":"Rate limit reached"}}`, func(err error) bool { return errors.Is(err, completion.TooManyRequests) }},
		{400, `{"error":{"message":"maximum context length is 10 tokens","code":"context_length_exceeded"}}`, func(err error) bool { return errors.Is(err, completion.ContextWindowExceeded) }},
		{500, `{"error":{"message":"kaputt"}}`, func(err error) bool { return strings.Contains(err.Error(), "kaputt") }},
		{200, `{"error":{"message":"in-band failure"}}`, func(err error) bool { return strings.Contains(err.Error(), "in-band failure") }},
		{200, `{"choices":[]}`, func(err error) bool { return strings.Contains(err.Error(), "no choices") }},
	}

	for _, c := range cases {
		fs := newFakeServer(t, func(w http.ResponseWriter, r *http.Request, body map[string]any) {
			writeJSON(w, c.status, c.body)
		})
		_, err := fs.provider("").completions.Complete(context.Background(), nil, userOpts("hi"))
		if err == nil || !c.check(err) {
			t.Errorf("%d %s: unexpected error %v", c.status, c.body, err)
		}
	}
}

func TestComplete_MaxTokensFallback(t *testing.T) {
	fs := newFakeServer(t, func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		if _, ok := body["max_tokens"]; ok {
			writeJSON(w, 400, `{"error":{"message":"Unsupported parameter: 'max_tokens' is not supported with this model. Use 'max_completion_tokens' instead.","type":"invalid_request_error","param":"max_tokens","code":"unsupported_parameter"}}`)
			return
		}
		writeJSON(w, 200, `{"choices":[{"finish_reason":"stop","message":{"content":"ok"}}]}`)
	})

	p := fs.provider("")
	opts := userOpts("hi")
	opts.MaxTokens = 10

	if _, err := p.completions.Complete(context.Background(), nil, opts); err != nil {
		t.Fatal(err)
	}
	if fs.calls() != 2 || fs.body(1)["max_completion_tokens"] != float64(10) {
		t.Fatalf("expected a retry with max_completion_tokens, got %d calls, %v", fs.calls(), fs.body(fs.calls()-1))
	}

	// the decision is remembered
	if _, err := p.completions.Complete(context.Background(), nil, opts); err != nil {
		t.Fatal(err)
	}
	if fs.calls() != 3 {
		t.Errorf("expected no further retry, got %d calls", fs.calls())
	}
}

func TestStream(t *testing.T) {
	fs := newFakeServer(t, func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		writeSSE(w,
			`{"choices":[{"index":0,"delta":{"role":"assistant","content":""}}]}`,
			`{"choices":[{"index":0,"delta":{"reasoning_content":"thinking..."}}]}`,
			`{"choices":[{"index":0,"delta":{"content":"Hel"}}]}`,
			`{"choices":[{"index":0,"delta":{"content":"lo"}}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"weather","arguments":""}}]}}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"city\":"}}]}}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"id":"call_2","type":"function","function":{"name":"time","arguments":"{}"}}]}}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"Bremen\"}"}}]}}]}`,
			`{"choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}`,
			`{"choices":[],"usage":{"prompt_tokens":11,"completion_tokens":7}}`,
			`[DONE]`,
		)
	})

	var text strings.Builder
	var calls []completion.ToolCall
	var done completion.Delta
	for d, err := range fs.provider("").completions.Stream(context.Background(), nil, userOpts("hi")) {
		if err != nil {
			t.Fatal(err)
		}
		text.WriteString(d.TextDelta)
		if d.ToolCall.IsSome() {
			calls = append(calls, d.ToolCall.Unwrap())
		}
		if d.Done {
			done = d
		}
	}

	if text.String() != "Hello" {
		t.Errorf("unexpected text %q", text.String())
	}
	if len(calls) != 2 || calls[0].ID != "call_1" || string(calls[0].Arguments) != `{"city":"Bremen"}` || calls[1].Name != "time" {
		t.Errorf("unexpected tool calls: %+v", calls)
	}
	if !done.Done || done.StopReason != completion.StopToolUse || done.Usage.Unwrap().InputTokens != 11 || done.Usage.Unwrap().OutputTokens != 7 {
		t.Errorf("unexpected final delta: %+v", done)
	}

	body := fs.body(0)
	if body["stream"] != true || body["stream_options"].(map[string]any)["include_usage"] != true {
		t.Errorf("expected stream with include_usage, got %v", body)
	}
}

func TestStream_CompleteToolCallsWithoutIndex(t *testing.T) {
	// some servers send each call complete, without index and a "stop" finish reason
	fs := newFakeServer(t, func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		writeSSE(w,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"id":"a","function":{"name":"x","arguments":"{\"n\":1}"}},{"id":"b","function":{"name":"y","arguments":{"n":2}}}]}}]}`,
			`{"choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
		)
	})

	var calls []completion.ToolCall
	var stop completion.StopReason
	for d, err := range fs.provider("").completions.Stream(context.Background(), nil, userOpts("hi")) {
		if err != nil {
			t.Fatal(err)
		}
		if d.ToolCall.IsSome() {
			calls = append(calls, d.ToolCall.Unwrap())
		}
		if d.Done {
			stop = d.StopReason
		}
	}

	if len(calls) != 2 || string(calls[0].Arguments) != `{"n":1}` || string(calls[1].Arguments) != `{"n":2}` || stop != completion.StopToolUse {
		t.Errorf("unexpected calls %+v / %s", calls, stop)
	}
}

func TestStream_StreamOptionsUnsupported(t *testing.T) {
	fs := newFakeServer(t, func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		if _, ok := body["stream_options"]; ok {
			writeJSON(w, 400, `{"error":{"message":"Unrecognized request argument supplied: stream_options"}}`)
			return
		}
		writeSSE(w, `{"choices":[{"index":0,"delta":{"content":"ok"},"finish_reason":"stop"}]}`, `[DONE]`)
	})

	p := fs.provider("")
	for range 2 {
		var text string
		for d, err := range p.completions.Stream(context.Background(), nil, userOpts("hi")) {
			if err != nil {
				t.Fatal(err)
			}
			text += d.TextDelta
		}
		if text != "ok" {
			t.Errorf("unexpected text %q", text)
		}
	}

	if fs.calls() != 3 {
		t.Errorf("expected one rejected and two accepted calls, got %d", fs.calls())
	}
}

func TestStream_Errors(t *testing.T) {
	cases := []struct {
		name  string
		serve func(w http.ResponseWriter)
		check func(error) bool
	}{
		{"429", func(w http.ResponseWriter) { writeJSON(w, 429, `{"error":{"message":"slow"}}`) }, func(err error) bool { return errors.Is(err, completion.TooManyRequests) }},
		{"in-stream", func(w http.ResponseWriter) {
			writeSSE(w, `{"choices":[{"index":0,"delta":{"content":"a"}}]}`, `{"error":{"message":"model crashed"}}`)
		}, func(err error) bool { return strings.Contains(err.Error(), "model crashed") }},
		{"garbage", func(w http.ResponseWriter) { writeSSE(w, `not json`) }, func(err error) bool { return strings.Contains(err.Error(), "decode stream chunk") }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := newFakeServer(t, func(w http.ResponseWriter, r *http.Request, body map[string]any) { c.serve(w) })
			var got error
			for _, err := range fs.provider("").completions.Stream(context.Background(), nil, userOpts("hi")) {
				if err != nil {
					got = err
				}
			}
			if got == nil || !c.check(got) {
				t.Errorf("unexpected error %v", got)
			}
		})
	}
}

func TestStream_ConsumerBreak(t *testing.T) {
	fs := newFakeServer(t, func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		writeSSE(w,
			`{"choices":[{"index":0,"delta":{"content":"a"}}]}`,
			`{"choices":[{"index":0,"delta":{"content":"b"}}]}`,
			`{"choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
		)
	})

	n := 0
	for d, err := range fs.provider("").completions.Stream(context.Background(), nil, userOpts("hi")) {
		if err != nil {
			t.Fatal(err)
		}
		n++
		if d.TextDelta != "a" {
			t.Errorf("unexpected delta %+v", d)
		}
		break
	}
	if n != 1 {
		t.Errorf("expected exactly one delta, got %d", n)
	}
}

func TestModels(t *testing.T) {
	fs := newFakeServer(t, func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		if r.URL.Path != "/v1/models" || r.Method != http.MethodGet {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, `{"object":"list","data":[
			{"id":"gpt-4o","object":"model","owned_by":"openai"},
			{"id":"text-embedding-3-small","object":"model"},
			{"id":"a-local-model","object":"model","owned_by":"library"},
			{"id":"whisper-1","object":"model"}
		]}`)
	})

	var ids []string
	for m, err := range fs.provider("").completions.Models(nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, string(m.ID))
		if m.Name != string(m.ID) {
			t.Errorf("unexpected name %q", m.Name)
		}
	}

	if strings.Join(ids, ",") != "a-local-model,gpt-4o" {
		t.Errorf("unexpected models %v", ids)
	}

	failing := newFakeServer(t, func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		writeJSON(w, 401, `{"error":{"message":"bad key"}}`)
	})
	for _, err := range failing.provider("").completions.Models(nil) {
		if err == nil || !strings.Contains(err.Error(), "bad key") {
			t.Errorf("unexpected error %v", err)
		}
	}
}

// TestRun drives the agentic loop of the completion package through a tool call roundtrip.
func TestRun(t *testing.T) {
	fs := newFakeServer(t, func(w http.ResponseWriter, r *http.Request, body map[string]any) {
		msgs := body["messages"].([]any)
		last := msgs[len(msgs)-1].(map[string]any)
		if last["role"] == "tool" {
			if last["tool_call_id"] != "call_1" || !strings.Contains(last["content"].(string), "42") {
				t.Errorf("unexpected tool message %v", last)
			}
			writeJSON(w, 200, `{"choices":[{"finish_reason":"stop","message":{"content":"The answer is 42."}}]}`)
			return
		}
		writeJSON(w, 200, `{"choices":[{"finish_reason":"tool_calls","message":{"content":null,"tool_calls":[{"id":"call_1","type":"function","function":{"name":"answer","arguments":"{\"question\":\"everything\"}"}}]}}]}`)
	})

	type In struct {
		Question string `json:"question"`
	}
	type Out struct {
		Answer int `json:"answer"`
	}

	tool := completion.NewTool("answer", "answers a question", func(in In) (Out, error) {
		if in.Question != "everything" {
			return Out{}, fmt.Errorf("unexpected question %q", in.Question)
		}
		return Out{Answer: 42}, nil
	})

	res, history, err := completion.Run(nil, fs.provider("").completions, completion.RunOptions{
		Options: userOpts("what is the answer?"),
		Tools:   []completion.Tool{tool},
	})
	if err != nil {
		t.Fatal(err)
	}

	if res.StopReason != completion.StopEndTurn || res.Message.Content[0].(completion.Text).Text != "The answer is 42." {
		t.Errorf("unexpected result %+v", res)
	}
	if len(history) != 4 {
		t.Errorf("expected user, assistant, tool result and answer in history, got %d", len(history))
	}
	if tools := fs.body(0)["tools"].([]any); len(tools) != 1 {
		t.Errorf("expected the tool to be advertised, got %v", tools)
	}
}

// blockingServer answers only once the client gave up (or after a generous safety timeout), and reports
// whether the request was aborted by the client. With streamFirst, it sends a first chunk before blocking, so
// cancellation of a running stream is exercised.
func blockingServer(t *testing.T, streamFirst bool) (*httptest.Server, <-chan struct{}) {
	t.Helper()
	aborted := make(chan struct{}, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The server notices a closed connection only once the request body was consumed.
		_, _ = io.Copy(io.Discard, r.Body)

		if streamFirst {
			writeSSE(w, `{"choices":[{"index":0,"delta":{"content":"a"}}]}`)
		}

		select {
		case <-r.Context().Done():
			aborted <- struct{}{}
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, aborted
}

func TestCancellationAbortsRequest(t *testing.T) {
	calls := map[string]struct {
		streamFirst bool
		call        func(ctx context.Context, p *openaiProvider) error
	}{
		"complete": {false, func(ctx context.Context, p *openaiProvider) error {
			_, err := p.completions.Complete(ctx, nil, userOpts("hi"))
			return err
		}},
		"stream": {false, func(ctx context.Context, p *openaiProvider) error {
			for _, err := range p.completions.Stream(ctx, nil, userOpts("hi")) {
				if err != nil {
					return err
				}
			}
			return nil
		}},
		"running stream": {true, func(ctx context.Context, p *openaiProvider) error {
			for _, err := range p.completions.Stream(ctx, nil, userOpts("hi")) {
				if err != nil {
					return err
				}
			}
			return nil
		}},
	}

	for name, c := range calls {
		t.Run(name, func(t *testing.T) {
			srv, aborted := blockingServer(t, c.streamFirst)
			p := newProvider("test", Settings{BaseURL: srv.URL})

			ctx, cancel := context.WithCancel(context.Background())
			time.AfterFunc(50*time.Millisecond, cancel)

			start := time.Now()
			err := c.call(ctx, p)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("expected context.Canceled, got %v", err)
			}
			if d := time.Since(start); d > 5*time.Second {
				t.Fatalf("cancellation took %s; the request was not aborted", d)
			}

			select {
			case <-aborted:
			case <-time.After(5 * time.Second):
				t.Fatal("the server never saw the request being aborted")
			}
		})
	}
}
