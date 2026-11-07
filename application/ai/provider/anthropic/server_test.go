// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
)

// searchResponse is the example answer of the documentation of the web search tool.
const searchResponse = `{
  "id": "msg_a930390d3a", "type": "message", "role": "assistant", "model": "claude-test",
  "content": [
    {"type": "text", "text": "I'll search for when Claude Shannon was born."},
    {"type": "server_tool_use", "id": "srvtoolu_01WYG3ziw53XMcoyKL4XcZmE", "name": "web_search", "input": {"query": "claude shannon birth date"}},
    {"type": "web_search_tool_result", "tool_use_id": "srvtoolu_01WYG3ziw53XMcoyKL4XcZmE", "content": [
      {"type": "web_search_result", "url": "https://en.wikipedia.org/wiki/Claude_Shannon", "title": "Claude Shannon - Wikipedia", "encrypted_content": "EqgfCioIARgBIiQ3YTAwMjY1Mi1mZjM5", "page_age": "April 30, 2025"}
    ]},
    {"text": "Based on the search results, ", "type": "text"},
    {"text": "Claude Shannon was born on April 30, 1916, in Petoskey, Michigan", "type": "text", "citations": [
      {"type": "web_search_result_location", "url": "https://en.wikipedia.org/wiki/Claude_Shannon", "title": "Claude Shannon - Wikipedia", "encrypted_index": "Eo8BCioIAhgBIiQyYjQ0OWJmZi1lNm..", "cited_text": "Claude Elwood Shannon (April 30, 1916 – February 24, 2001) was an American mathematician"}
    ]}
  ],
  "stop_reason": "end_turn",
  "usage": {"input_tokens": 6039, "output_tokens": 931, "server_tool_use": {"web_search_requests": 1}}
}`

// fetchResponse is the example answer of the documentation of the web fetch tool, whose result has an object
// as content.
const fetchResponse = `{
  "id": "msg_b", "type": "message", "role": "assistant", "model": "claude-test",
  "content": [
    {"type": "server_tool_use", "id": "srvtoolu_01234567890abcdef", "name": "web_fetch", "input": {"url": "https://example.com/article"}},
    {"type": "web_fetch_tool_result", "tool_use_id": "srvtoolu_01234567890abcdef", "content": {
      "type": "web_fetch_result", "url": "https://example.com/article",
      "content": {"type": "document", "source": {"type": "text", "media_type": "text/plain", "data": "Full text content of the article..."}, "title": "Article Title", "citations": {"enabled": true}},
      "retrieved_at": "2025-08-25T10:30:00Z"
    }},
    {"type": "text", "text": "The article says it."}
  ],
  "stop_reason": "end_turn",
  "usage": {"input_tokens": 25039, "output_tokens": 931, "server_tool_use": {"web_fetch_requests": 1}}
}`

func decodeResponse(t *testing.T, body string) completion.Result {
	t.Helper()
	var resp apiResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("cannot decode the response: %v", err)
	}
	return fromAPIResponse(resp)
}

// jsonEqual compares two JSON documents regardless of formatting and key order.
func jsonEqual(t *testing.T, a, b []byte) bool {
	t.Helper()
	var x, y any
	if err := json.Unmarshal(a, &x); err != nil {
		t.Fatalf("invalid json %s: %v", a, err)
	}
	if err := json.Unmarshal(b, &y); err != nil {
		t.Fatalf("invalid json %s: %v", b, err)
	}
	xa, _ := json.Marshal(x)
	ya, _ := json.Marshal(y)
	return bytes.Equal(xa, ya)
}

func TestServerBlocksAreKeptAndReplayedUnchanged(t *testing.T) {
	for name, body := range map[string]string{"search": searchResponse, "fetch": fetchResponse} {
		t.Run(name, func(t *testing.T) {
			res := decodeResponse(t, body)

			var original struct {
				Content []json.RawMessage `json:"content"`
			}
			_ = json.Unmarshal([]byte(body), &original)
			if len(res.Message.Content) != len(original.Content) {
				t.Fatalf("expected %d blocks, got %d: %#v", len(original.Content), len(res.Message.Content), res.Message.Content)
			}

			// the turn survives the persistence of a session
			buf, err := json.Marshal(res.Message)
			if err != nil {
				t.Fatal(err)
			}
			var restored completion.Message
			if err := json.Unmarshal(buf, &restored); err != nil {
				t.Fatal(err)
			}

			p := &anthropicProvider{cfg: Settings{DisablePromptCache: true}}
			req, err := p.buildRequest(completion.Options{
				Model:       "claude-test",
				ServerTools: completion.WebResearchTools(),
				Messages: []completion.Message{
					{Role: completion.User, Content: []completion.Content{completion.Text{Text: "question"}}},
					restored,
					{Role: completion.User, Content: []completion.Content{completion.Text{Text: "thanks"}}},
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			sent, err := json.Marshal(req.Messages[1].Content)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := json.Marshal(original.Content)
			if !jsonEqual(t, sent, want) {
				t.Fatalf("the turn changed on its way back:\nsent %s\nwant %s", sent, want)
			}
		})
	}
}

func TestServerBlockViews(t *testing.T) {
	res := decodeResponse(t, searchResponse)
	if res.Usage.WebSearches != 1 || res.Usage.InputTokens != 6039 {
		t.Fatalf("unexpected usage %+v", res.Usage)
	}

	call, ok := res.Message.Content[1].(completion.ServerBlock).Call()
	if !ok || call.Kind != completion.WebSearch || call.Query != "claude shannon birth date" {
		t.Fatalf("unexpected call %+v", call)
	}

	hits, ok := res.Message.Content[2].(completion.ServerBlock).Result()
	if !ok || hits.CallID != call.ID || len(hits.Sources) != 1 || hits.Sources[0].Title != "Claude Shannon - Wikipedia" {
		t.Fatalf("unexpected result %+v", hits)
	}

	if cited := res.Message.Content[4].(completion.Text); len(cited.Citations) == 0 {
		t.Fatal("the citations of the text were dropped")
	}

	fetched := decodeResponse(t, fetchResponse)
	page, ok := fetched.Message.Content[1].(completion.ServerBlock).Result()
	if !ok || page.Kind != completion.WebFetch || page.Text != "Full text content of the article..." || page.Sources[0].URL != "https://example.com/article" {
		t.Fatalf("unexpected fetch %+v", page)
	}
	if fetched.Usage.WebFetches != 1 {
		t.Fatalf("unexpected usage %+v", fetched.Usage)
	}

	flat := completion.FlattenServerBlocks([]completion.Message{fetched.Message})
	for _, c := range flat[0].Content {
		if _, ok := c.(completion.Text); !ok {
			t.Fatalf("a flattened history must contain text only, got %T", c)
		}
	}
	if !strings.Contains(flat[0].Content[1].(completion.Text).Text, "Full text content") {
		t.Fatalf("the fetched text got lost: %+v", flat[0].Content[1])
	}
}

func TestServerToolDefinitions(t *testing.T) {
	p := &anthropicProvider{cfg: Settings{}}
	opts := baseOpts()
	opts.ServerTools = []completion.ServerTool{
		{Kind: completion.WebSearch, MaxUses: 5, AllowedDomains: []string{"example.com"}},
		{Kind: completion.WebFetch, MaxUses: 10, MaxContentTokens: 25000},
	}

	req, err := p.buildRequest(opts)
	if err != nil {
		t.Fatal(err)
	}

	buf, _ := json.Marshal(req.Tools)
	want := `[
		{"type":"web_search_20250305","name":"web_search","max_uses":5,"allowed_domains":["example.com"]},
		{"type":"web_fetch_20250910","name":"web_fetch","max_uses":10,"max_content_tokens":25000},
		{"name":"a","input_schema":{}},
		{"name":"b","input_schema":{},"cache_control":{"type":"ephemeral"}}
	]`
	if !jsonEqual(t, buf, []byte(want)) {
		t.Fatalf("unexpected tools %s", buf)
	}

	// without custom tools, no server tool is marked as cache breakpoint
	opts.Tools = nil
	req, _ = p.buildRequest(opts)
	for _, tool := range req.Tools {
		if tool.CacheControl != nil {
			t.Fatalf("server tool %s carries a cache breakpoint", tool.Name)
		}
	}

	opts.ServerTools = []completion.ServerTool{{Kind: "code_interpreter"}}
	if _, err := p.buildRequest(opts); err == nil {
		t.Fatal("an unknown server tool must be refused")
	}
}

func TestServerBlocksNeverCarryCacheBreakpoints(t *testing.T) {
	fetched := decodeResponse(t, fetchResponse)
	// the turn ends with the result block, which would be the breakpoint of the history otherwise
	fetched.Message.Content = fetched.Message.Content[:2]

	p := &anthropicProvider{cfg: Settings{}}
	req, err := p.buildRequest(completion.Options{
		Model: "claude-test",
		Messages: []completion.Message{
			{Role: completion.User, Content: []completion.Content{completion.Text{Text: "read https://example.com/article"}}},
			fetched.Message,
			{Role: completion.User, Content: []completion.Content{completion.Text{Text: "and now?"}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	buf, _ := json.Marshal(req.Messages)
	if strings.Count(string(buf), "cache_control") != 1 || !strings.Contains(string(buf), `"text":"read https://example.com/article","cache_control"`) &&
		!strings.Contains(string(buf), `"cache_control":{"type":"ephemeral"},"text":"read https://example.com/article"`) {
		t.Fatalf("the breakpoint must move to the text before the server blocks: %s", buf)
	}
}

func TestRefusedServerToolsAreDroppedForTheModel(t *testing.T) {
	refusals := map[string]string{
		// a gateway without a search names the tool, the other one stays
		"named": `the server tool web_search is not available on the models of openai: no search is configured`,
		// a refusal without a name drops all of them
		"unnamed": `server tools are not supported by this model`,
	}

	for name, refusal := range refusals {
		t.Run(name, func(t *testing.T) {
			var mu sync.Mutex
			var bodies []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				buf, _ := io.ReadAll(r.Body)
				mu.Lock()
				bodies = append(bodies, string(buf))
				n := len(bodies)
				mu.Unlock()

				w.Header().Set("Content-Type", "application/json")
				if n == 1 {
					w.WriteHeader(http.StatusBadRequest)
					msg, _ := json.Marshal(refusal)
					_, _ = w.Write([]byte(`{"type":"error","error":{"type":"invalid_request_error","message":` + string(msg) + `}}`))
					return
				}
				_, _ = w.Write([]byte(`{"id":"msg","type":"message","role":"assistant","model":"m","content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`))
			}))
			t.Cleanup(srv.Close)

			p := NewProviderAt("test", Settings{}, Endpoint{BaseURL: srv.URL + "/v1/"}).(*anthropicProvider)

			fetched := decodeResponse(t, fetchResponse)
			opts := completion.Options{
				Model:       "m",
				ServerTools: completion.WebResearchTools(),
				Messages: []completion.Message{
					{Role: completion.User, Content: []completion.Content{completion.Text{Text: "read https://example.com/article"}}},
					fetched.Message,
					{Role: completion.User, Content: []completion.Content{completion.Text{Text: "and now?"}}},
				},
			}

			for range 2 {
				res, err := p.completions.Complete(context.Background(), nil, opts)
				if err != nil {
					t.Fatal(err)
				}
				if res.Message.Content[0].(completion.Text).Text != "ok" {
					t.Fatalf("unexpected answer %+v", res)
				}
			}

			mu.Lock()
			defer mu.Unlock()
			if len(bodies) != 3 {
				t.Fatalf("expected the refused request, its retry and one more request, got %d", len(bodies))
			}

			for _, body := range bodies[1:] {
				if strings.Contains(body, webSearchType) {
					t.Fatalf("a request after the refusal still offers the search: %s", body)
				}
				if !strings.Contains(body, "Full text content of the article") {
					t.Fatalf("the fetched text must stay readable: %s", body)
				}

				keepsFetch := strings.Contains(body, webFetchType) && strings.Contains(body, "web_fetch_tool_result")
				if keepsFetch != (name == "named") {
					t.Fatalf("the fetch must stay with its blocks only if the refusal named the search alone: %s", body)
				}
			}
		})
	}
}

// A stream carries the result of a server tool in one content_block_start, whose content is an object. It must
// not break the stream, and the call of the server tool is no tool call of the client.
func TestStreamWithServerBlocks(t *testing.T) {
	events := []string{
		`event: message_start` + "\n" + `data: {"type":"message_start","message":{"id":"msg","type":"message","usage":{"input_tokens":5}}}`,
		`event: content_block_start` + "\n" + `data: {"type":"content_block_start","index":0,"content_block":{"type":"server_tool_use","id":"srvtoolu_1","name":"web_fetch"}}`,
		`event: content_block_delta` + "\n" + `data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"url\":\"https://example.com\"}"}}`,
		`event: content_block_stop` + "\n" + `data: {"type":"content_block_stop","index":0}`,
		`event: content_block_start` + "\n" + `data: {"type":"content_block_start","index":1,"content_block":{"type":"web_fetch_tool_result","tool_use_id":"srvtoolu_1","content":{"type":"web_fetch_result","url":"https://example.com","content":{"type":"document","source":{"type":"text","media_type":"text/plain","data":"x"}}}}}`,
		`event: content_block_stop` + "\n" + `data: {"type":"content_block_stop","index":1}`,
		`event: content_block_start` + "\n" + `data: {"type":"content_block_start","index":2,"content_block":{"type":"text","text":""}}`,
		`event: content_block_delta` + "\n" + `data: {"type":"content_block_delta","index":2,"delta":{"type":"text_delta","text":"done"}}`,
		`event: content_block_stop` + "\n" + `data: {"type":"content_block_stop","index":2}`,
		`event: message_delta` + "\n" + `data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":3,"server_tool_use":{"web_fetch_requests":1}}}`,
		`event: message_stop` + "\n" + `data: {"type":"message_stop"}`,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		for _, ev := range events {
			_, _ = io.WriteString(w, ev+"\n\n")
		}
	}))
	t.Cleanup(srv.Close)

	p := NewProviderAt("test", Settings{}, Endpoint{BaseURL: srv.URL + "/v1/"}).(*anthropicProvider)
	var text string
	for d, err := range p.completions.Stream(context.Background(), nil, completion.Options{Model: "m", ServerTools: completion.WebResearchTools(), Messages: []completion.Message{
		{Role: completion.User, Content: []completion.Content{completion.Text{Text: "read https://example.com"}}},
	}}) {
		if err != nil {
			t.Fatal(err)
		}
		if d.ToolCall.IsSome() {
			t.Fatalf("a server tool must not become a tool call of the client: %+v", d.ToolCall.Unwrap())
		}
		text += d.TextDelta
		if d.Done && d.StopReason != completion.StopEndTurn {
			t.Fatalf("unexpected stop %s", d.StopReason)
		}
	}

	if text != "done" {
		t.Fatalf("unexpected text %q", text)
	}
}
