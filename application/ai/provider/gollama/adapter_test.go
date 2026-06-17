// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
)

func weatherTool() completion.ToolDef {
	return completion.ToolDef{
		Name:        "get_weather",
		Description: "Get the weather for a city",
		Schema:      json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`),
	}
}

func TestDetectFamily(t *testing.T) {
	cases := []struct {
		name string
		meta ggufMetadata
		want family
	}{
		{"chatml template", ggufMetadata{chatTemplate: "{% for m %}<|im_start|>{{m.role}}"}, familyChatML},
		{"llama3 template", ggufMetadata{chatTemplate: "<|start_header_id|>system<|end_header_id|>"}, familyLlama3},
		{"mistral template", ggufMetadata{chatTemplate: "[INST] {{ x }} [/INST]"}, familyMistral},
		{"qwen arch", ggufMetadata{architecture: "qwen2"}, familyChatML},
		{"llama3 by name", ggufMetadata{architecture: "llama", name: "Meta Llama 3.2 3B"}, familyLlama3},
		{"llama2 spm -> mistral", ggufMetadata{architecture: "llama", name: "Llama 2 7B", tokenizerModel: "llama"}, familyMistral},
		{"unknown -> chatml", ggufMetadata{architecture: "phi3"}, familyChatML},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detectFamily(tc.meta); got != tc.want {
				t.Fatalf("detectFamily = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestChatMLRenderPrompt(t *testing.T) {
	opts := completion.Options{
		System: "You are helpful.",
		Tools:  []completion.ToolDef{weatherTool()},
		Messages: []completion.Message{
			{Role: completion.User, Content: []completion.Content{completion.Text{Text: "Weather in NYC?"}}},
		},
	}
	got, err := chatmlAdapter{}.renderPrompt(opts)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"<|im_start|>system\nYou are helpful.",
		"<tools>\n",
		`"name":"get_weather"`,
		"</tools>",
		"<|im_start|>user\nWeather in NYC?<|im_end|>",
		"<|im_start|>assistant\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt missing %q in:\n%s", want, got)
		}
	}
	if !strings.HasSuffix(got, "<|im_start|>assistant\n") {
		t.Errorf("prompt must end ready for assistant generation, got tail:\n%q", got[max(0, len(got)-40):])
	}
}

func TestChatMLParseToolCall(t *testing.T) {
	gen := "Let me check.<tool_call>\n{\"name\": \"get_weather\", \"arguments\": {\"city\": \"NYC\"}}\n</tool_call>"
	contents, reason := chatmlAdapter{}.parse(gen)
	if reason != completion.StopToolUse {
		t.Fatalf("reason = %q, want tool_use", reason)
	}
	assertText(t, contents, "Let me check.")
	call := assertSingleToolCall(t, contents)
	if call.Name != "get_weather" {
		t.Errorf("call name = %q", call.Name)
	}
	if got := compactJSON(t, call.Arguments); got != `{"city":"NYC"}` {
		t.Errorf("args = %s", got)
	}
}

func TestChatMLParseThinking(t *testing.T) {
	contents, reason := chatmlAdapter{}.parse("<think>hmm</think>The answer is 42.")
	if reason != completion.StopEndTurn {
		t.Fatalf("reason = %q, want end_turn", reason)
	}
	if len(contents) != 2 {
		t.Fatalf("want thinking+text, got %d blocks: %#v", len(contents), contents)
	}
	if th, ok := contents[0].(completion.Thinking); !ok || th.Text != "hmm" {
		t.Errorf("first block = %#v, want Thinking{hmm}", contents[0])
	}
	assertText(t, contents, "The answer is 42.")
}

func TestLlama3ParseBareJSON(t *testing.T) {
	contents, reason := llama3Adapter{}.parse(`{"name": "get_weather", "parameters": {"city": "NYC"}}`)
	if reason != completion.StopToolUse {
		t.Fatalf("reason = %q, want tool_use", reason)
	}
	call := assertSingleToolCall(t, contents)
	if call.Name != "get_weather" || compactJSON(t, call.Arguments) != `{"city":"NYC"}` {
		t.Errorf("call = %+v args=%s", call, call.Arguments)
	}
}

func TestLlama3ParseFunctionTag(t *testing.T) {
	contents, reason := llama3Adapter{}.parse(`<function=get_weather>{"city": "NYC"}</function>`)
	if reason != completion.StopToolUse {
		t.Fatalf("reason = %q, want tool_use", reason)
	}
	call := assertSingleToolCall(t, contents)
	if call.Name != "get_weather" {
		t.Errorf("call name = %q", call.Name)
	}
}

func TestLlama3ParsePlainText(t *testing.T) {
	contents, reason := llama3Adapter{}.parse("Hello there!")
	if reason != completion.StopEndTurn {
		t.Fatalf("reason = %q, want end_turn", reason)
	}
	assertText(t, contents, "Hello there!")
}

func TestLlama3RenderToolResultUsesIPython(t *testing.T) {
	opts := completion.Options{
		Messages: []completion.Message{
			{Role: completion.User, Content: []completion.Content{completion.Text{Text: "Weather?"}}},
			{Role: completion.Assistant, Content: []completion.Content{completion.ToolCall{ID: "call_1", Name: "get_weather", Arguments: json.RawMessage(`{"city":"NYC"}`)}}},
			{Role: completion.User, Content: []completion.Content{completion.ToolResult{ToolCallID: "call_1", Content: []completion.Content{completion.Text{Text: "sunny"}}}}},
		},
	}
	got, err := llama3Adapter{}.renderPrompt(opts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "<|start_header_id|>ipython<|end_header_id|>\n\nsunny<|eot_id|>") {
		t.Errorf("missing ipython tool result turn in:\n%s", got)
	}
}

func TestMistralRenderToolsBeforeLastUser(t *testing.T) {
	opts := completion.Options{
		System: "Sys",
		Tools:  []completion.ToolDef{weatherTool()},
		Messages: []completion.Message{
			{Role: completion.User, Content: []completion.Content{completion.Text{Text: "Hi"}}},
		},
	}
	got, err := mistralAdapter{}.renderPrompt(opts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "[AVAILABLE_TOOLS]") || !strings.Contains(got, "[/AVAILABLE_TOOLS]") {
		t.Errorf("missing available tools block in:\n%s", got)
	}
	if !strings.Contains(got, "[INST] Sys\n\nHi [/INST]") {
		t.Errorf("system not merged into first INST in:\n%s", got)
	}
	if strings.Index(got, "[AVAILABLE_TOOLS]") > strings.Index(got, "[INST]") {
		t.Errorf("available tools must precede the last user INST in:\n%s", got)
	}
}

func TestMistralParseToolCalls(t *testing.T) {
	contents, reason := mistralAdapter{}.parse(`[TOOL_CALLS][{"name": "get_weather", "arguments": {"city": "NYC"}}]`)
	if reason != completion.StopToolUse {
		t.Fatalf("reason = %q, want tool_use", reason)
	}
	call := assertSingleToolCall(t, contents)
	if call.Name != "get_weather" {
		t.Errorf("call name = %q", call.Name)
	}
}

// ----- helpers -----

func assertText(t *testing.T, contents []completion.Content, want string) {
	t.Helper()
	for _, c := range contents {
		if txt, ok := c.(completion.Text); ok {
			if txt.Text == want {
				return
			}
			t.Fatalf("text = %q, want %q", txt.Text, want)
		}
	}
	t.Fatalf("no text block found, want %q in %#v", want, contents)
}

func assertSingleToolCall(t *testing.T, contents []completion.Content) completion.ToolCall {
	t.Helper()
	var calls []completion.ToolCall
	for _, c := range contents {
		if tc, ok := c.(completion.ToolCall); ok {
			calls = append(calls, tc)
		}
	}
	if len(calls) != 1 {
		t.Fatalf("want exactly 1 tool call, got %d in %#v", len(calls), contents)
	}
	if calls[0].ID == "" {
		t.Errorf("tool call must have a generated id")
	}
	return calls[0]
}

func compactJSON(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		t.Fatalf("compact %s: %v", raw, err)
	}
	return buf.String()
}
