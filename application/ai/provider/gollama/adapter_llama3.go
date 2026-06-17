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

// llama3Adapter renders the Llama 3.x header prompt format (<|start_header_id|>role<|end_header_id|>) and its
// JSON tool-calling convention. Tool results are fed back through the dedicated "ipython" role turn. The model
// may answer a tool call as a bare JSON object {"name":..., "parameters":...}, optionally prefixed with the
// <|python_tag|> marker, or using the <function=NAME>{...}</function> form.
type llama3Adapter struct{}

const llama3ToolPreamble = "You have access to the following functions. To call a function, respond with a " +
	"single JSON object (and nothing else) of the form " +
	`{"name": <function-name>, "parameters": <arguments-json-object>}` + ". " +
	"Only call functions from the list below; otherwise answer normally."

func (llama3Adapter) name() string { return "llama3" }

func (llama3Adapter) renderPrompt(opts completion.Options) (string, error) {
	var b strings.Builder

	sys := strings.TrimSpace(opts.System)
	if sys != "" || len(opts.Tools) > 0 {
		b.WriteString("<|start_header_id|>system<|end_header_id|>\n\n")
		if sys != "" {
			b.WriteString(sys)
		}
		if len(opts.Tools) > 0 {
			if sys != "" {
				b.WriteString("\n\n")
			}
			b.WriteString(llama3ToolPreamble)
			b.WriteString("\n\n")
			for _, t := range opts.Tools {
				line, err := json.Marshal(toFnDef(t))
				if err != nil {
					return "", err
				}
				b.Write(line)
				b.WriteByte('\n')
			}
		}
		b.WriteString("<|eot_id|>")
	}

	for _, m := range opts.Messages {
		switch m.Role {
		case completion.User:
			results := toolResultsOf(m.Content)
			if len(results) > 0 {
				for _, tr := range results {
					b.WriteString("<|start_header_id|>ipython<|end_header_id|>\n\n")
					b.WriteString(textOfContent(tr.Content))
					b.WriteString("<|eot_id|>")
				}
				continue
			}
			b.WriteString("<|start_header_id|>user<|end_header_id|>\n\n")
			b.WriteString(textOfContent(m.Content))
			b.WriteString("<|eot_id|>")
		case completion.Assistant:
			b.WriteString("<|start_header_id|>assistant<|end_header_id|>\n\n")
			b.WriteString(renderLlama3Assistant(m.Content))
			b.WriteString("<|eot_id|>")
		}
	}

	b.WriteString("<|start_header_id|>assistant<|end_header_id|>\n\n")
	return b.String(), nil
}

func renderLlama3Assistant(content []completion.Content) string {
	var b strings.Builder
	for _, c := range content {
		switch v := c.(type) {
		case completion.Text:
			b.WriteString(v.Text)
		case completion.ToolCall:
			obj, _ := json.Marshal(struct {
				Name       string          `json:"name"`
				Parameters json.RawMessage `json:"parameters"`
			}{Name: v.Name, Parameters: rawOrEmpty(v.Arguments)})
			b.Write(obj)
		}
	}
	return b.String()
}

func (llama3Adapter) stopStrings() []string {
	return []string{"<|eot_id|>", "<|eom_id|>", "<|end_of_text|>"}
}

func (llama3Adapter) toolMarkers() []string {
	return []string{"<|python_tag|>", "<function="}
}

func (llama3Adapter) parse(generated string) ([]completion.Content, completion.StopReason) {
	s := strings.TrimSpace(generated)
	s = strings.TrimSpace(strings.TrimPrefix(s, "<|python_tag|>"))

	if strings.HasPrefix(s, "<function=") {
		if calls := parseFunctionTagCalls(s); len(calls) > 0 {
			return asContents(calls), completion.StopToolUse
		}
	}

	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		if calls := parseJSONToolCalls(s); len(calls) > 0 {
			return asContents(calls), completion.StopToolUse
		}
	}

	if s == "" {
		return nil, completion.StopEndTurn
	}
	return []completion.Content{completion.Text{Text: s}}, completion.StopEndTurn
}

// parseFunctionTagCalls decodes one or more <function=NAME>{json-args}</function> blocks.
func parseFunctionTagCalls(s string) []completion.ToolCall {
	var out []completion.ToolCall
	for {
		i := strings.Index(s, "<function=")
		if i < 0 {
			break
		}
		s = s[i+len("<function="):]
		gt := strings.IndexByte(s, '>')
		if gt < 0 {
			break
		}
		name := strings.TrimSpace(s[:gt])
		s = s[gt+1:]
		end := strings.Index(s, "</function>")
		args := s
		if end >= 0 {
			args = s[:end]
			s = s[end+len("</function>"):]
		} else {
			s = ""
		}
		if name == "" {
			continue
		}
		out = append(out, completion.ToolCall{
			ID:        newToolCallID(),
			Name:      name,
			Arguments: rawOrEmpty(json.RawMessage(strings.TrimSpace(args))),
		})
	}
	return out
}
