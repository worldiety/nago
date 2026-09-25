// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"encoding/json"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
)

func askHistory(args, answer string) []completion.Message {
	return []completion.Message{
		{Role: completion.User, Content: []completion.Content{completion.Text{Text: "plan my trip"}}},
		{Role: completion.Assistant, Content: []completion.Content{
			completion.ToolCall{ID: "a1", Name: askUserToolName, Arguments: json.RawMessage(args)},
		}},
		{Role: completion.User, Content: []completion.Content{
			completion.ToolResult{ToolCallID: "a1", Content: []completion.Content{completion.Text{Text: answer}}},
		}},
		{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: "ok, Berlin"}}},
	}
}

// TestRenderHistory_AskUserStaysVisible: after answering, the clarifying question and the answer must remain
// part of the visible conversation instead of collapsing into a tool hint.
func TestRenderHistory_AskUserStaysVisible(t *testing.T) {
	views := renderHistory(nil, askHistory(`{"question":"where to?","options":["Berlin","Rome"]}`, `{"answer":"Berlin"}`))
	// user prompt, question, answer, final reply
	if len(views) != 4 {
		t.Fatalf("expected 4 bubbles, got %d", len(views))
	}
}

func TestRenderHistory_AskUserFallsBackOnBrokenPayload(t *testing.T) {
	views := renderHistory(nil, askHistory(`not json`, `{"answer":"Berlin"}`))
	// user prompt, tool hint, final reply; the answer of an unrecognised call stays hidden
	if len(views) != 3 {
		t.Fatalf("expected 3 views, got %d", len(views))
	}
}

func TestRenderHistory_OtherToolResultsStayHidden(t *testing.T) {
	h := askHistory(`{"question":"q"}`, `{"answer":"a"}`)
	h[1].Content[0] = completion.ToolCall{ID: "a1", Name: "search", Arguments: json.RawMessage(`{}`)}
	views := renderHistory(nil, h)
	// user prompt, tool hint, final reply
	if len(views) != 3 {
		t.Fatalf("expected 3 views, got %d", len(views))
	}
}

func TestAskPayloadParsing(t *testing.T) {
	q, ok := askQuestion(completion.ToolCall{Arguments: json.RawMessage(`{"question":"where?"}`)})
	if !ok || q != "where?" {
		t.Fatalf("question = %q, %v", q, ok)
	}
	a, ok := askAnswer(completion.ToolResult{Content: []completion.Content{completion.Text{Text: `{"answer":"Rome"}`}}})
	if !ok || a != "Rome" {
		t.Fatalf("answer = %q, %v", a, ok)
	}
	if _, ok := askAnswer(completion.ToolResult{Content: []completion.Content{completion.Text{Text: `{"answer":""}`}}}); ok {
		t.Fatalf("empty answer must not render")
	}
}
