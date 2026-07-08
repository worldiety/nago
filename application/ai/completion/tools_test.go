// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"encoding/json"
	"iter"
	"testing"

	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
)

type addIn struct {
	A int `json:"a" desc:"first summand"`
	B int `json:"b"`
}

type addOut struct {
	Sum int `json:"sum"`
}

func TestNewTool_Schema(t *testing.T) {
	tool := NewTool("add", "adds two integers", func(in addIn) (addOut, error) {
		return addOut{Sum: in.A + in.B}, nil
	})

	if tool.Def.Name != "add" {
		t.Fatalf("unexpected name: %q", tool.Def.Name)
	}

	var schema map[string]any
	if err := json.Unmarshal(tool.Def.Schema, &schema); err != nil {
		t.Fatalf("schema not valid json: %v", err)
	}

	if schema["type"] != "object" {
		t.Fatalf("expected object schema, got %v", schema["type"])
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("missing properties: %v", schema)
	}

	a, ok := props["a"].(map[string]any)
	if !ok || a["type"] != "integer" || a["description"] != "first summand" {
		t.Fatalf("unexpected schema for a: %v", props["a"])
	}

	required, ok := schema["required"].([]any)
	if !ok || len(required) != 2 {
		t.Fatalf("expected both fields required, got %v", schema["required"])
	}
}

func TestNewTool_Invoke(t *testing.T) {
	tool := NewTool("add", "", func(in addIn) (addOut, error) {
		return addOut{Sum: in.A + in.B}, nil
	})

	out, err := tool.Invoke(json.RawMessage(`{"a":2,"b":40}`))
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}

	if string(out) != `{"sum":42}` {
		t.Fatalf("unexpected result: %s", out)
	}
}

func TestNewContentTool_Schema(t *testing.T) {
	tool := NewContentTool("render", "renders something", func(in addIn) ([]Content, error) {
		return nil, nil
	})

	if tool.InvokeContent == nil {
		t.Fatal("expected InvokeContent to be set")
	}
	if tool.Invoke != nil {
		t.Fatal("expected Invoke to be nil for a content tool")
	}

	var schema map[string]any
	if err := json.Unmarshal(tool.Def.Schema, &schema); err != nil {
		t.Fatalf("schema not valid json: %v", err)
	}
	if schema["type"] != "object" {
		t.Fatalf("expected object schema, got %v", schema["type"])
	}
}

func TestNewContentTool_InvokeContent(t *testing.T) {
	tool := NewContentTool("render", "", func(in addIn) ([]Content, error) {
		return []Content{
			Text{Text: "here is your file"},
			FileRef{File: "file-1", MimeType: "application/pdf"},
		}, nil
	})

	blocks, err := tool.InvokeContent(json.RawMessage(`{"a":1,"b":2}`))
	if err != nil {
		t.Fatalf("invoke content failed: %v", err)
	}

	if len(blocks) != 2 {
		t.Fatalf("expected 2 content blocks, got %d", len(blocks))
	}
	if _, ok := blocks[0].(Text); !ok {
		t.Fatalf("expected first block to be Text, got %T", blocks[0])
	}
	fr, ok := blocks[1].(FileRef)
	if !ok || fr.File != "file-1" {
		t.Fatalf("unexpected second block: %#v", blocks[1])
	}
}

// TestRun_ContentToolResult verifies that a content tool's blocks land verbatim as ToolResult.Content in the
// history the loop builds.
func TestRun_ContentToolResult(t *testing.T) {
	tool := NewContentTool("makefile", "", func(in addIn) ([]Content, error) {
		return []Content{
			Text{Text: "done"},
			FileRef{File: "file-42", MimeType: "application/pdf"},
		}, nil
	})

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					ToolCall{ID: "1", Name: "makefile", Arguments: json.RawMessage(`{"a":1,"b":2}`)},
				}},
				StopReason: StopToolUse,
			},
			{
				Message:    Message{Role: Assistant, Content: []Content{Text{Text: "the file is ready"}}},
				StopReason: StopEndTurn,
			},
		},
	}

	_, history, err := Run(nil, fake, RunOptions{
		Options: Options{
			Messages: []Message{{Role: User, Content: []Content{Text{Text: "make a file"}}}},
		},
		Tools: []Tool{tool},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	tr, ok := history[2].Content[0].(ToolResult)
	if !ok || tr.ToolCallID != "1" || tr.IsError {
		t.Fatalf("unexpected tool result: %#v", history[2].Content[0])
	}
	if len(tr.Content) != 2 {
		t.Fatalf("expected 2 content blocks in tool result, got %d: %#v", len(tr.Content), tr.Content)
	}
	if _, ok := tr.Content[1].(FileRef); !ok {
		t.Fatalf("expected FileRef in tool result, got %T", tr.Content[1])
	}
}

// fakeCompletions returns the queued results in order, so we can simulate a tool-use turn followed by a
// final answer.
type fakeCompletions struct {
	results []Result
	calls   int
}

func (f *fakeCompletions) Models(auth.Subject) iter.Seq2[model.Model, error] { return nil }

func (f *fakeCompletions) Complete(_ auth.Subject, _ Options) (Result, error) {
	r := f.results[f.calls]
	f.calls++
	return r, nil
}

func (f *fakeCompletions) Stream(auth.Subject, Options) iter.Seq2[Delta, error] { return nil }

func TestRun_ExecutesToolLoop(t *testing.T) {
	tool := NewTool("add", "", func(in addIn) (addOut, error) {
		return addOut{Sum: in.A + in.B}, nil
	})

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					ToolCall{ID: "1", Name: "add", Arguments: json.RawMessage(`{"a":2,"b":40}`)},
				}},
				StopReason: StopToolUse,
			},
			{
				Message: Message{Role: Assistant, Content: []Content{
					Text{Text: "the sum is 42"},
				}},
				StopReason: StopEndTurn,
			},
		},
	}

	res, history, err := Run(nil, fake, RunOptions{
		Options: Options{
			Messages: []Message{{Role: User, Content: []Content{Text{Text: "add 2 and 40"}}}},
		},
		Tools: []Tool{tool},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if res.StopReason != StopEndTurn {
		t.Fatalf("unexpected stop reason: %v", res.StopReason)
	}

	// user prompt + assistant tool_use + user tool_result + assistant final = 4 messages
	if len(history) != 4 {
		t.Fatalf("unexpected history length %d: %+v", len(history), history)
	}

	toolResultMsg := history[2]
	if toolResultMsg.Role != User || len(toolResultMsg.Content) != 1 {
		t.Fatalf("expected tool result user message, got %+v", toolResultMsg)
	}

	tr, ok := toolResultMsg.Content[0].(ToolResult)
	if !ok || tr.ToolCallID != "1" || tr.IsError {
		t.Fatalf("unexpected tool result: %+v", toolResultMsg.Content[0])
	}
}

// hasDanglingToolUse reports whether any assistant message contains a ToolCall that is not answered by a
// ToolResult with the same ID in the immediately following message. This is exactly the condition the
// Anthropic API rejects with a 400.
func hasDanglingToolUse(history []Message) bool {
	for i, m := range history {
		var ids []string
		for _, c := range m.Content {
			if call, ok := c.(ToolCall); ok {
				ids = append(ids, call.ID)
			}
		}
		if len(ids) == 0 {
			continue
		}

		answered := map[string]bool{}
		if i+1 < len(history) {
			for _, c := range history[i+1].Content {
				if tr, ok := c.(ToolResult); ok {
					answered[tr.ToolCallID] = true
				}
			}
		}
		for _, id := range ids {
			if !answered[id] {
				return true
			}
		}
	}
	return false
}

// TestRun_DropsTruncatedToolUse reproduces the 400 from the bug report: the model emitted tool_use blocks
// but the turn was cut off (stop_reason == max_tokens). The loop must not persist a dangling tool_use.
func TestRun_DropsTruncatedToolUse(t *testing.T) {
	tool := NewTool("add", "", func(in addIn) (addOut, error) {
		return addOut{Sum: in.A + in.B}, nil
	})

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					Text{Text: "let me compute"},
					ToolCall{ID: "1", Name: "add", Arguments: json.RawMessage(`{"a":2,"b":40}`)},
					ToolCall{ID: "2", Name: "add", Arguments: json.RawMessage(`{"a":1,`)}, // truncated args
				}},
				StopReason: StopMaxTokens,
			},
		},
	}

	res, history, err := Run(nil, fake, RunOptions{
		Options: Options{
			Messages: []Message{{Role: User, Content: []Content{Text{Text: "add some numbers"}}}},
		},
		Tools: []Tool{tool},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if res.StopReason != StopMaxTokens {
		t.Fatalf("unexpected stop reason: %v", res.StopReason)
	}

	if hasDanglingToolUse(history) {
		t.Fatalf("history contains a tool_use without a matching tool_result: %+v", history)
	}

	// user prompt + assistant text (tool_use stripped) = 2 messages; no tool was executed.
	if len(history) != 2 {
		t.Fatalf("unexpected history length %d: %+v", len(history), history)
	}

	last := history[1]
	if last.Role != Assistant || len(last.Content) != 1 {
		t.Fatalf("expected single-block assistant message, got %+v", last)
	}
	if _, ok := last.Content[0].(Text); !ok {
		t.Fatalf("expected remaining text block, got %T", last.Content[0])
	}
}

// TestRun_TruncatedToolUseOnlyDropsMessage verifies that an aborted turn consisting solely of tool_use
// blocks (no text/thinking) is dropped entirely rather than persisted as an empty assistant message.
func TestRun_TruncatedToolUseOnlyDropsMessage(t *testing.T) {
	tool := NewTool("add", "", func(in addIn) (addOut, error) {
		return addOut{Sum: in.A + in.B}, nil
	})

	fake := &fakeCompletions{
		results: []Result{
			{
				Message: Message{Role: Assistant, Content: []Content{
					ToolCall{ID: "1", Name: "add", Arguments: json.RawMessage(`{"a":2,`)},
				}},
				StopReason: StopMaxTokens,
			},
		},
	}

	_, history, err := Run(nil, fake, RunOptions{
		Options: Options{
			Messages: []Message{{Role: User, Content: []Content{Text{Text: "add"}}}},
		},
		Tools: []Tool{tool},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if hasDanglingToolUse(history) {
		t.Fatalf("history contains a dangling tool_use: %+v", history)
	}

	// Only the original user prompt remains.
	if len(history) != 1 {
		t.Fatalf("unexpected history length %d: %+v", len(history), history)
	}
}
