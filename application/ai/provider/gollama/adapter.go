// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"

	"go.wdy.de/nago/application/ai/completion"
)

// adapter isolates everything that differs between GGUF model families: how the stateless completion request
// is rendered into a single prompt string (including tool definitions and prior tool results), which strings
// terminate an assistant turn, and how the generated text is parsed back into content blocks (text, thinking,
// tool calls).
//
// Detokenization is handled separately by [pieceDecoder] because it depends on the tokenizer, not the chat
// format.
type adapter interface {
	// name identifies the family for diagnostics.
	name() string
	// renderPrompt builds the full prompt. It must NOT include the leading BOS token: the engine tokenizes
	// with addSpecial=true so llama.cpp inserts the model-specific BOS (if any).
	renderPrompt(opts completion.Options) (string, error)
	// stopStrings lists markers that end an assistant turn (in addition to the model's EOS token).
	stopStrings() []string
	// toolMarkers lists substrings that signal the start of a tool-call section. While streaming they switch
	// off incremental text emission so the raw tool-call syntax is never surfaced as visible assistant text;
	// the full output is parsed once generation completes.
	toolMarkers() []string
	// parse converts the raw generated assistant text into content blocks and a stop reason.
	parse(generated string) ([]completion.Content, completion.StopReason)
}

// selectAdapter resolves the family adapter for a model. An explicit catalog override wins; otherwise the
// family is detected from the GGUF metadata.
func selectAdapter(override family, meta ggufMetadata) adapter {
	fam := override
	if fam == familyAuto {
		fam = detectFamily(meta)
	}

	switch fam {
	case familyLlama3:
		return llama3Adapter{}
	case familyMistral:
		return mistralAdapter{}
	case familyChatML:
		return chatmlAdapter{}
	default:
		return chatmlAdapter{}
	}
}

// detectFamily infers the family from GGUF metadata. The chat template fingerprint is the most reliable
// signal; the architecture and name are used as fallbacks.
func detectFamily(meta ggufMetadata) family {
	tpl := meta.chatTemplate
	switch {
	case strings.Contains(tpl, "<|im_start|>"):
		return familyChatML
	case strings.Contains(tpl, "<|start_header_id|>"):
		return familyLlama3
	case strings.Contains(tpl, "[AVAILABLE_TOOLS]"), strings.Contains(tpl, "[INST]"):
		return familyMistral
	}

	switch strings.ToLower(meta.architecture) {
	case "qwen2", "qwen3", "qwen2moe", "qwen3moe":
		return familyChatML
	case "llama":
		name := strings.ToLower(meta.name)
		if strings.Contains(name, "llama 3") || strings.Contains(name, "llama-3") || strings.Contains(name, "llama3") {
			return familyLlama3
		}
		if meta.isByteLevelBPE() {
			return familyLlama3
		}
		return familyMistral
	case "mistral", "mixtral":
		return familyMistral
	}

	// ChatML is the most widely compatible default for current instruct GGUFs.
	return familyChatML
}

// ----- shared tool encoding/decoding helpers -----

// fnDef is the OpenAI-style function-tool description shared by the Hermes (ChatML), Llama 3 and Mistral
// encodings.
type fnDef struct {
	Type     string     `json:"type"`
	Function fnInnerDef `json:"function"`
}

type fnInnerDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

func toFnDef(t completion.ToolDef) fnDef {
	schema := t.Schema
	if len(schema) == 0 {
		schema = json.RawMessage("{}")
	}
	return fnDef{
		Type: "function",
		Function: fnInnerDef{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  schema,
		},
	}
}

// jsonToolCall is the lenient on-the-wire shape a model emits for a tool call. Both "arguments" (Hermes,
// Mistral) and "parameters" (Llama 3) are accepted.
type jsonToolCall struct {
	Name       string          `json:"name"`
	Arguments  json.RawMessage `json:"arguments"`
	Parameters json.RawMessage `json:"parameters"`
}

func (c jsonToolCall) args() json.RawMessage {
	if len(c.Arguments) > 0 {
		return c.Arguments
	}
	if len(c.Parameters) > 0 {
		return c.Parameters
	}
	return json.RawMessage("{}")
}

func (c jsonToolCall) toToolCall() completion.ToolCall {
	return completion.ToolCall{
		ID:        newToolCallID(),
		Name:      c.Name,
		Arguments: c.args(),
	}
}

// newToolCallID mints a unique id used to correlate a generated tool call with the [completion.ToolResult]
// the caller feeds back. Local models do not emit ids of their own.
func newToolCallID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "call_0"
	}
	return "call_" + hex.EncodeToString(b[:])
}

// rawOrEmpty returns raw if it carries JSON, otherwise an empty JSON object.
func rawOrEmpty(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage("{}")
	}
	return raw
}

// parseJSONToolCall decodes a single JSON object into a tool call. It returns false if the text is not a JSON
// object carrying a non-empty name.
func parseJSONToolCall(s string) (completion.ToolCall, bool) {
	var c jsonToolCall
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &c); err != nil {
		return completion.ToolCall{}, false
	}
	if c.Name == "" {
		return completion.ToolCall{}, false
	}
	return c.toToolCall(), true
}

// parseJSONToolCalls decodes one or more tool calls from text that is either a single object, a JSON array of
// objects, or several whitespace/`;`-separated objects (all forms emitted by local models). Only entries with
// a non-empty name are returned.
func parseJSONToolCalls(s string) []completion.ToolCall {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// JSON array of calls.
	if strings.HasPrefix(s, "[") {
		var arr []jsonToolCall
		if err := json.Unmarshal([]byte(s), &arr); err == nil {
			var out []completion.ToolCall
			for _, c := range arr {
				if c.Name != "" {
					out = append(out, c.toToolCall())
				}
			}
			return out
		}
	}

	// One or more concatenated objects, optionally separated by ';' or whitespace.
	var out []completion.ToolCall
	dec := json.NewDecoder(strings.NewReader(s))
	for {
		var c jsonToolCall
		if err := dec.Decode(&c); err != nil {
			break
		}
		if c.Name != "" {
			out = append(out, c.toToolCall())
		}
	}
	return out
}

// toolResultsOf returns the tool-result blocks carried by a message's content.
func toolResultsOf(content []completion.Content) []completion.ToolResult {
	var out []completion.ToolResult
	for _, c := range content {
		if tr, ok := c.(completion.ToolResult); ok {
			out = append(out, tr)
		}
	}
	return out
}

// toolCallsOf returns the tool-call blocks carried by a message's content.
func toolCallsOf(content []completion.Content) []completion.ToolCall {
	var out []completion.ToolCall
	for _, c := range content {
		if tc, ok := c.(completion.ToolCall); ok {
			out = append(out, tc)
		}
	}
	return out
}

// textOfContent concatenates the plain text blocks of a message's content.
func textOfContent(content []completion.Content) string {
	var sb strings.Builder
	for _, c := range content {
		if t, ok := c.(completion.Text); ok {
			sb.WriteString(t.Text)
		}
	}
	return sb.String()
}

// asContents widens tool calls to content blocks.
func asContents(calls []completion.ToolCall) []completion.Content {
	out := make([]completion.Content, 0, len(calls))
	for _, c := range calls {
		out = append(out, c)
	}
	return out
}

// splitOut extracts every substring enclosed by open/close markers, returning the collected inner segments
// and the text with those segments (and markers) removed.
func splitOut(s, open, close string) (inner []string, rest string) {
	var b strings.Builder
	for {
		i := strings.Index(s, open)
		if i < 0 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:i])
		s = s[i+len(open):]
		j := strings.Index(s, close)
		if j < 0 {
			// unterminated: treat the remainder as inner content
			inner = append(inner, s)
			s = ""
			break
		}
		inner = append(inner, s[:j])
		s = s[j+len(close):]
	}
	return inner, b.String()
}
