// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/auth"
)

func serverCall(id, name, input string) ServerBlock {
	return ServerBlock{Raw: json.RawMessage(`{"type":"server_tool_use","id":"` + id + `","name":"` + name + `","input":` + input + `}`)}
}

func fetchResult(id, url, text string) ServerBlock {
	return ServerBlock{Raw: json.RawMessage(`{"type":"web_fetch_tool_result","tool_use_id":"` + id + `","content":{"type":"web_fetch_result","url":"` + url + `","content":{"type":"document","source":{"type":"text","media_type":"text/plain","data":"` + text + `"},"title":"Title"},"retrieved_at":"2025-08-25T10:30:00Z"}}`)}
}

func TestServerBlockJSONRoundTrip(t *testing.T) {
	msg := Message{Role: Assistant, Content: []Content{
		serverCall("srv1", "web_fetch", `{"url":"https://example.com"}`),
		fetchResult("srv1", "https://example.com", "hello"),
		Text{Text: "cited", Citations: json.RawMessage(`[{"type":"char_location","document_index":0}]`)},
	}}

	buf, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	var got Message
	if err := json.Unmarshal(buf, &got); err != nil {
		t.Fatal(err)
	}

	for i := range msg.Content {
		a, _ := json.Marshal(Message{Content: []Content{msg.Content[i]}})
		b, _ := json.Marshal(Message{Content: []Content{got.Content[i]}})
		if !bytes.Equal(a, b) {
			t.Fatalf("block %d changed: %s != %s", i, a, b)
		}
	}

	if res, ok := got.Content[1].(ServerBlock).Result(); !ok || res.Text != "hello" || res.CallID != "srv1" {
		t.Fatalf("unexpected result %+v", res)
	}
}

func TestServerErrorResult(t *testing.T) {
	b := ServerBlock{Raw: json.RawMessage(`{"type":"web_search_tool_result","tool_use_id":"s","content":{"type":"web_search_tool_result_error","error_code":"max_uses_exceeded"}}`)}
	res, ok := b.Result()
	if !ok || res.Kind != WebSearch || res.ErrorCode != "max_uses_exceeded" {
		t.Fatalf("unexpected result %+v", res)
	}
}

// A model which calls a server tool together with a client tool gets the result of the server tool with its
// next answer. The turn in between must consist of tool results only, so the attachment of a file-providing tool
// joins the next turn of tool results.
func TestRun_MixedTurnKeepsTheResultTurnClean(t *testing.T) {
	open := NewOpenFileTool("open_file", "opens a file", func(in openIn) (OpenedFile, error) {
		return OpenedFile{Name: "report.pdf", MimeType: file.PDF, Open: func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("%PDF-1.7"))), nil
		}}, nil
	})
	add := NewTool("add", "adds", func(in addIn) (addOut, error) { return addOut{Sum: in.A + in.B}, nil })

	fake := &fakeCompletions{results: []Result{
		assistantText(StopToolUse,
			serverCall("srv1", "web_fetch", `{"url":"https://example.com"}`),
			ToolCall{ID: "c1", Name: "open_file", Arguments: json.RawMessage(`{"fid":"abc"}`)},
		),
		assistantText(StopToolUse,
			fetchResult("srv1", "https://example.com", "hello"),
			ToolCall{ID: "c2", Name: "add", Arguments: json.RawMessage(`{"a":1,"b":2}`)},
		),
		assistantText(StopEndTurn, Text{Text: "done"}),
	}}

	_, history, err := Run(nil, fake, RunOptions{
		Options: Options{Messages: userMsg("go"), ServerTools: WebResearchTools()},
		Tools:   []Tool{open, add},
		FileUploader: func(auth.Subject, OpenedFile) (file.ID, error) {
			return "file-1", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(fake.reqs[0].ServerTools) != 2 {
		t.Fatalf("the server tools must reach the provider, got %+v", fake.reqs[0].ServerTools)
	}

	// user, mixed turn, tool results only, deferred result with a call, tool results with the attachment, answer
	if len(history) != 6 {
		t.Fatalf("unexpected history %+v", history)
	}

	for _, c := range history[2].Content {
		if _, ok := c.(ToolResult); !ok {
			t.Fatalf("the turn after a pending server call may only carry tool results, got %T", c)
		}
	}

	var media bool
	for _, c := range history[4].Content {
		if _, ok := c.(Media); ok {
			media = true
		}
	}
	if !media {
		t.Fatalf("the attachment must join the next turn of tool results: %+v", history[4])
	}
}

func TestRun_TruncatedTurnDropsPendingServerCalls(t *testing.T) {
	fake := &fakeCompletions{results: []Result{
		assistantText(StopMaxTokens, Text{Text: "let me look"}, serverCall("srv1", "web_search", `{"query":"x"}`)),
		assistantText(StopEndTurn, Text{Text: "answer"}),
	}}

	_, history, err := Run(nil, fake, RunOptions{Options: Options{Messages: userMsg("q"), MaxTokens: maxEscalatedTokens, ServerTools: WebResearchTools()}})
	if err != nil {
		t.Fatal(err)
	}

	for _, m := range history {
		if pendingServerCall(m) {
			t.Fatalf("a truncated turn kept a server call nobody answers: %+v", history)
		}
	}
}

func TestRun_ToolNamedLikeServerToolIsRefused(t *testing.T) {
	tool := NewTool("web_search", "my own search", func(in addIn) (addOut, error) { return addOut{}, nil })
	_, _, err := Run(nil, &fakeCompletions{results: []Result{assistantText(StopEndTurn, Text{Text: "x"})}}, RunOptions{
		Options: Options{Messages: userMsg("q"), ServerTools: WebResearchTools()},
		Tools:   []Tool{tool},
	})
	if err == nil {
		t.Fatal("a tool must not collide with a server tool")
	}
}

func TestCompactionKeepsServerPairsTogether(t *testing.T) {
	history := []Message{
		{Role: User, Content: []Content{Text{Text: "q"}}},
		{Role: Assistant, Content: []Content{serverCall("srv1", "web_fetch", `{"url":"https://example.com"}`), ToolCall{ID: "c1", Name: "add", Arguments: json.RawMessage(`{}`)}}},
		{Role: User, Content: []Content{ToolResult{ToolCallID: "c1", Content: []Content{Text{Text: "3"}}}}},
		{Role: Assistant, Content: []Content{fetchResult("srv1", "https://example.com", "hello"), Text{Text: "done"}}},
		{Role: User, Content: []Content{Text{Text: "next"}}},
	}

	// the tail must neither start with the tool results nor with the deferred result
	if start := splitIndex(history, 3); start != 4 {
		t.Fatalf("unexpected split %d", start)
	}

	flat := truncateLargest(history)
	if HasServerBlocks(flat) {
		t.Fatalf("the last resort must flatten every server block: %+v", flat)
	}
	if runeLen(flat) >= runeLen(history) {
		t.Fatal("the last resort must shrink the history")
	}
}
