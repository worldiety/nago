// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"encoding/json"
	"strings"

	"go.wdy.de/nago/application/ai/completion"
)

// chatmlAdapter renders the ChatML prompt format used by Qwen, Hermes and many other current instruct GGUFs,
// with Hermes-style tool calling: tool definitions are advertised inside a <tools> block in the system turn
// and the model answers with <tool_call>{...}</tool_call> blocks. Optional <think>...</think> reasoning is
// surfaced as [completion.Thinking].
type chatmlAdapter struct{}

const (
	chatmlToolsPreamble = "# Tools\n\nYou may call one or more functions to assist with the user query.\n\n" +
		"You are provided with function signatures within <tools></tools> XML tags:"

	chatmlToolCallInstruction = "For each function call, return a json object with function name and arguments " +
		"within <tool_call></tool_call> XML tags:\n<tool_call>\n" +
		`{"name": <function-name>, "arguments": <args-json-object>}` + "\n</tool_call>"
)

func (chatmlAdapter) name() string { return "chatml" }

func (chatmlAdapter) renderPrompt(opts completion.Options) (string, error) {
	var b strings.Builder

	sys := strings.TrimSpace(opts.System)
	if sys != "" || len(opts.Tools) > 0 {
		b.WriteString("<|im_start|>system\n")
		if sys != "" {
			b.WriteString(sys)
		}
		if len(opts.Tools) > 0 {
			if sys != "" {
				b.WriteString("\n\n")
			}
			b.WriteString(chatmlToolsPreamble)
			b.WriteString("\n<tools>\n")
			for _, t := range opts.Tools {
				line, err := json.Marshal(toFnDef(t))
				if err != nil {
					return "", err
				}
				b.Write(line)
				b.WriteByte('\n')
			}
			b.WriteString("</tools>\n\n")
			b.WriteString(chatmlToolCallInstruction)
		}
		b.WriteString("<|im_end|>\n")
	}

	for _, m := range opts.Messages {
		switch m.Role {
		case completion.User:
			b.WriteString("<|im_start|>user\n")
			b.WriteString(renderChatMLUser(m.Content))
			b.WriteString("<|im_end|>\n")
		case completion.Assistant:
			b.WriteString("<|im_start|>assistant\n")
			b.WriteString(renderChatMLAssistant(m.Content))
			b.WriteString("<|im_end|>\n")
		}
	}

	b.WriteString("<|im_start|>assistant\n")
	return b.String(), nil
}

func renderChatMLUser(content []completion.Content) string {
	var parts []string
	for _, c := range content {
		switch v := c.(type) {
		case completion.Text:
			if v.Text != "" {
				parts = append(parts, v.Text)
			}
		case completion.ToolResult:
			parts = append(parts, "<tool_response>\n"+textOfContent(v.Content)+"\n</tool_response>")
		}
	}
	return strings.Join(parts, "\n")
}

func renderChatMLAssistant(content []completion.Content) string {
	var b strings.Builder
	for _, c := range content {
		switch v := c.(type) {
		case completion.Text:
			b.WriteString(v.Text)
		case completion.ToolCall:
			obj, _ := json.Marshal(struct {
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			}{Name: v.Name, Arguments: rawOrEmpty(v.Arguments)})
			b.WriteString("<tool_call>\n")
			b.Write(obj)
			b.WriteString("\n</tool_call>")
		}
	}
	return b.String()
}

func (chatmlAdapter) stopStrings() []string {
	return []string{"<|im_end|>", "<|endoftext|>"}
}

func (chatmlAdapter) toolMarkers() []string {
	return []string{"<tool_call>"}
}

func (chatmlAdapter) parse(generated string) ([]completion.Content, completion.StopReason) {
	var out []completion.Content

	thinks, afterThink := splitOut(generated, "<think>", "</think>")
	for _, t := range thinks {
		if tt := strings.TrimSpace(t); tt != "" {
			out = append(out, completion.Thinking{Text: tt})
		}
	}

	rawCalls, rest := splitOut(afterThink, "<tool_call>", "</tool_call>")

	if t := strings.TrimSpace(rest); t != "" {
		out = append(out, completion.Text{Text: t})
	}

	var calls []completion.ToolCall
	for _, rc := range rawCalls {
		if tc, ok := parseJSONToolCall(rc); ok {
			calls = append(calls, tc)
		}
	}
	out = append(out, asContents(calls)...)

	if len(calls) > 0 {
		return out, completion.StopToolUse
	}
	return out, completion.StopEndTurn
}
