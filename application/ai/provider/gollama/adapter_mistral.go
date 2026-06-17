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

// mistralAdapter renders the Mistral / Mixtral instruct format: user turns are wrapped in [INST] ... [/INST],
// assistant turns are terminated by </s>, tool definitions are advertised in an [AVAILABLE_TOOLS] block placed
// directly before the final user instruction, tool calls are emitted as [TOOL_CALLS][{...}] and tool results
// are fed back inside [TOOL_RESULTS] ... [/TOOL_RESULTS]. The system prompt is merged into the first user
// instruction, as Mistral has no dedicated system role.
type mistralAdapter struct{}

func (mistralAdapter) name() string { return "mistral" }

func (mistralAdapter) renderPrompt(opts completion.Options) (string, error) {
	var b strings.Builder

	lastInst := -1
	for i, m := range opts.Messages {
		if m.Role == completion.User && len(toolResultsOf(m.Content)) == 0 {
			lastInst = i
		}
	}

	sys := strings.TrimSpace(opts.System)
	systemInjected := false

	for i, m := range opts.Messages {
		switch m.Role {
		case completion.User:
			if results := toolResultsOf(m.Content); len(results) > 0 {
				for _, tr := range results {
					payload, err := json.Marshal(struct {
						Content string `json:"content"`
					}{Content: textOfContent(tr.Content)})
					if err != nil {
						return "", err
					}
					b.WriteString("[TOOL_RESULTS]")
					b.Write(payload)
					b.WriteString("[/TOOL_RESULTS]")
				}
				continue
			}

			if i == lastInst && len(opts.Tools) > 0 {
				defs := make([]fnDef, 0, len(opts.Tools))
				for _, t := range opts.Tools {
					defs = append(defs, toFnDef(t))
				}
				arr, err := json.Marshal(defs)
				if err != nil {
					return "", err
				}
				b.WriteString("[AVAILABLE_TOOLS]")
				b.Write(arr)
				b.WriteString("[/AVAILABLE_TOOLS]")
			}

			b.WriteString("[INST] ")
			if !systemInjected && sys != "" {
				b.WriteString(sys)
				b.WriteString("\n\n")
				systemInjected = true
			}
			b.WriteString(textOfContent(m.Content))
			b.WriteString(" [/INST]")

		case completion.Assistant:
			if calls := toolCallsOf(m.Content); len(calls) > 0 {
				arr := make([]json.RawMessage, 0, len(calls))
				for _, c := range calls {
					obj, _ := json.Marshal(struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					}{Name: c.Name, Arguments: rawOrEmpty(c.Arguments)})
					arr = append(arr, obj)
				}
				joined, _ := json.Marshal(arr)
				b.WriteString("[TOOL_CALLS]")
				b.Write(joined)
				b.WriteString("</s>")
			} else {
				b.WriteString(textOfContent(m.Content))
				b.WriteString("</s>")
			}
		}
	}

	return b.String(), nil
}

func (mistralAdapter) stopStrings() []string {
	return []string{"</s>"}
}

func (mistralAdapter) toolMarkers() []string {
	return []string{"[TOOL_CALLS]"}
}

func (mistralAdapter) parse(generated string) ([]completion.Content, completion.StopReason) {
	if i := strings.Index(generated, "[TOOL_CALLS]"); i >= 0 {
		pre := strings.TrimSpace(generated[:i])
		rest := strings.TrimSpace(generated[i+len("[TOOL_CALLS]"):])

		var out []completion.Content
		if pre != "" {
			out = append(out, completion.Text{Text: pre})
		}
		calls := parseJSONToolCalls(rest)
		out = append(out, asContents(calls)...)
		if len(calls) > 0 {
			return out, completion.StopToolUse
		}
	}

	s := strings.TrimSpace(generated)
	if s == "" {
		return nil, completion.StopEndTurn
	}
	return []completion.Content{completion.Text{Text: s}}, completion.StopEndTurn
}
