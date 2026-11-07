// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"encoding/json"
	"strings"
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
	views := renderHistory(nil, askHistory(`{"question":"where to?","options":["Berlin","Rome"]}`, `{"answer":"Berlin"}`), historyView{})
	// user prompt, question, answer, final reply
	if len(views) != 4 {
		t.Fatalf("expected 4 bubbles, got %d", len(views))
	}
}

func TestRenderHistory_AskUserFallsBackOnBrokenPayload(t *testing.T) {
	views := renderHistory(nil, askHistory(`not json`, `{"answer":"Berlin"}`), historyView{})
	// user prompt, tool hint, final reply; the answer of an unrecognised call stays hidden
	if len(views) != 3 {
		t.Fatalf("expected 3 views, got %d", len(views))
	}
}

func TestRenderHistory_OtherToolResultsStayHidden(t *testing.T) {
	h := askHistory(`{"question":"q"}`, `{"answer":"a"}`)
	h[1].Content[0] = completion.ToolCall{ID: "a1", Name: "search", Arguments: json.RawMessage(`{}`)}
	views := renderHistory(nil, h, historyView{})
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

// TestRenderHistory_ConsecutiveToolCallsAreGrouped: a long agentic run must not flood the conversation, so
// consecutive tool calls, including the reasoning between and after them, become one collapsed section, which a
// visible answer ends. Reasoning without tool calls keeps its own section.
func TestRenderHistory_ConsecutiveToolCallsAreGrouped(t *testing.T) {
	call := func(id string) completion.Message {
		return completion.Message{Role: completion.Assistant, Content: []completion.Content{
			completion.Thinking{Text: "checking " + id},
			completion.ToolCall{ID: id, Name: "search", Arguments: json.RawMessage(`{}`)},
		}}
	}
	result := func(id string) completion.Message {
		return completion.Message{Role: completion.User, Content: []completion.Content{
			completion.ToolResult{ToolCallID: id, Content: []completion.Content{completion.Text{Text: "ok"}}},
		}}
	}

	history := []completion.Message{
		{Role: completion.User, Content: []completion.Content{completion.Text{Text: "question"}}},
		call("c1"), result("c1"), call("c2"), result("c2"), call("c3"), result("c3"),
		{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: "interim answer"}}},
		call("c4"), result("c4"),
		{Role: completion.Assistant, Content: []completion.Content{completion.Thinking{Text: "done"}, completion.Text{Text: "final answer"}}},
	}

	// question, group of three, interim answer, group of one with the trailing reasoning, final answer
	if views := renderHistory(nil, history, historyView{}); len(views) != 5 {
		t.Fatalf("expected 5 views, got %d", len(views))
	}

	// question, reasoning, answer
	plain := []completion.Message{
		history[0],
		{Role: completion.Assistant, Content: []completion.Content{completion.Thinking{Text: "easy"}, completion.Text{Text: "answer"}}},
	}
	if views := renderHistory(nil, plain, historyView{}); len(views) != 3 {
		t.Fatalf("expected 3 views, got %d", len(views))
	}
}

// TestRenderHistory_ServerToolsJoinTheToolSection: the calls of server tools are shown like tool calls, with
// what they found, and their results are no bubbles of their own.
func TestRenderHistory_ServerToolsJoinTheToolSection(t *testing.T) {
	h := []completion.Message{
		{Role: completion.User, Content: []completion.Content{completion.Text{Text: "who is Claude Shannon?"}}},
		{Role: completion.Assistant, Content: []completion.Content{
			completion.ServerBlock{Raw: json.RawMessage(`{"type":"server_tool_use","id":"s1","name":"web_search","input":{"query":"claude shannon"}}`)},
			completion.ServerBlock{Raw: json.RawMessage(`{"type":"web_search_tool_result","tool_use_id":"s1","content":[{"type":"web_search_result","url":"https://en.wikipedia.org/wiki/Claude_Shannon","title":"Claude Shannon","encrypted_content":"x"},{"type":"web_search_result","url":"javascript:alert(1)","title":"evil","encrypted_content":"y"}]}`)},
			completion.Text{Text: "a mathematician", Citations: json.RawMessage(`[]`)},
		}},
	}

	views := renderHistory(nil, h, historyView{})
	// user prompt, tool section, answer
	if len(views) != 3 {
		t.Fatalf("expected 3 views, got %d", len(views))
	}

	label := serverCallLabel(completion.ServerCall{Kind: completion.WebSearch, Query: "claude shannon"})
	if label != "Websuche „claude shannon“" {
		t.Fatalf("unexpected label %q", label)
	}
}

// TestJoinTexts: a cited answer comes in parts, which are one text again, with numbered links to the cited pages.
// A part never starts a line of its own, so that "1. Oktober" stays a date and does not become a list.
func TestJoinTexts(t *testing.T) {
	cite := func(urls ...string) json.RawMessage {
		var parts []string
		for _, u := range urls {
			parts = append(parts, `{"type":"web_search_result_location","url":"`+u+`","title":"t","encrypted_index":"x","cited_text":"c"}`)
		}
		return json.RawMessage("[" + strings.Join(parts, ",") + "]")
	}

	got := joinTexts([]completion.Text{
		{Text: "Eine Drittseite gibt den "},
		{Text: "1. Oktober 2026", Citations: cite("https://example.com/a", "https://example.com/b(1)")},
		{Text: " als Datum an, go.dev "},
		{Text: "den 19.10.2026.\n", Citations: cite("https://example.com/a", "javascript:alert(1)")},
		{Text: "Fertig."},
	})

	want := `Eine Drittseite gibt den 1. Oktober 2026 [\[1\]](https://example.com/a) [\[2\]](https://example.com/b%281%29) als Datum an, go.dev den 19.10.2026. [\[1\]](https://example.com/a)` + "\nFertig."
	if got != want {
		t.Fatalf("unexpected text\n got: %q\nwant: %q", got, want)
	}

	// texts without citations stay paragraphs of their own
	if got := joinTexts([]completion.Text{{Text: "a"}, {Text: "b"}}); got != "a\n\nb" {
		t.Fatalf("unexpected text %q", got)
	}

	h := []completion.Message{{Role: completion.Assistant, Content: []completion.Content{
		completion.Text{Text: "Go "}, completion.Text{Text: "1.27", Citations: cite("https://go.dev/doc")}, completion.Text{Text: " ist aktuell."},
	}}}
	if views := renderHistory(nil, h, historyView{}); len(views) != 1 {
		t.Fatalf("a cited answer must be one bubble, got %d views", len(views))
	}
}
