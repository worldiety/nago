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

