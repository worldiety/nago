// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package openai

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/model"
)

// buildRequest translates the stateless [completion.Options] into the Chat Completions wire request.
//
// The Anthropic-style content model of the completion package is mapped as follows:
//   - [completion.Options.System] becomes a leading system message,
//   - [completion.ToolResult] blocks of a user turn become role "tool" messages, which are placed directly
//     after the preceding assistant turn as the API demands, followed by the remaining user content,
//   - [completion.ToolCall] blocks of an assistant turn become its tool_calls,
//   - [completion.Thinking] and [completion.RedactedThinking] are dropped, because the API has no way to send
//     reasoning back (and some servers, e.g. DeepSeek, reject it).
func (p *openaiProvider) buildRequest(opts completion.Options) (apiRequest, error) {
	req := apiRequest{
		Model: string(opts.Model),
		Stop:  opts.StopSequences,
	}

	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = p.cfg.MaxTokens
	}
	if maxTokens > 0 {
		if p.useMaxCompletionTokens.Load() {
			req.MaxCompletionTokens = maxTokens
		} else {
			req.MaxTokens = maxTokens
		}
	}

	if opts.Temperature.IsSome() {
		v := opts.Temperature.Unwrap()
		req.Temperature = &v
	}

	if opts.TopP.IsSome() {
		v := opts.TopP.Unwrap()
		req.TopP = &v
	}

	if strings.TrimSpace(opts.System) != "" {
		req.Messages = append(req.Messages, apiMessage{Role: "system", Content: opts.System})
	}

	for _, m := range opts.Messages {
		msgs, err := toAPIMessages(m)
		if err != nil {
			return apiRequest{}, err
		}
		req.Messages = append(req.Messages, msgs...)
	}

	for _, t := range opts.Tools {
		req.Tools = append(req.Tools, toAPITool(t))
	}

	if len(req.Tools) > 0 {
		// tool_choice is rejected by OpenAI when no tools are given.
		req.ToolChoice = toAPIToolChoice(opts.ToolChoice)
	}

	return req, nil
}

func toAPITool(t completion.ToolDef) apiTool {
	fn := apiFunction{
		Name:        t.Name,
		Description: t.Description,
	}

	if s := strings.TrimSpace(string(t.Schema)); s != "" && s != "null" {
		fn.Parameters = t.Schema
	}

	return apiTool{Type: "function", Function: fn}
}

// toAPIToolChoice maps the tool choice: auto -> "auto", any -> "required", none -> "none" and a named tool to
// the function object form. The zero value omits the field, which means auto.
func toAPIToolChoice(tc completion.ToolChoice) any {
	switch {
	case tc.Name != "":
		return map[string]any{"type": "function", "function": map[string]string{"name": tc.Name}}
	case tc.Mode == "any":
		return "required"
	case tc.Mode == "none":
		return "none"
	case tc.Mode == "auto":
		return "auto"
	default:
		return nil
	}
}

// toAPIMessages translates a single turn into one or more chat messages.
func toAPIMessages(m completion.Message) ([]apiMessage, error) {
	switch m.Role {
	case completion.Assistant:
		msg, ok := toAPIAssistantMessage(m)
		if !ok {
			return nil, nil
		}
		return []apiMessage{msg}, nil
	case completion.User:
		return toAPIUserMessages(m)
	default:
		return nil, fmt.Errorf("unsupported message role %q", m.Role)
	}
}

// toAPIAssistantMessage joins all text blocks into the content and maps the tool calls. It reports false for a
// turn without any transferable content, which must be dropped because the API rejects it.
func toAPIAssistantMessage(m completion.Message) (apiMessage, bool) {
	var sb strings.Builder
	var calls []apiToolCall

	for _, c := range m.Content {
		switch v := c.(type) {
		case completion.Text:
			sb.WriteString(v.Text)
		case completion.ToolCall:
			args := strings.TrimSpace(string(v.Arguments))
			if args == "" || args == "null" {
				args = "{}"
			}
			calls = append(calls, apiToolCall{
				ID:       v.ID,
				Type:     "function",
				Function: apiToolCallFunction{Name: v.Name, Arguments: args},
			})
		case completion.Media:
			sb.WriteString(mediaPlaceholder(v))
		}
	}

	msg := apiMessage{Role: "assistant", ToolCalls: calls}
	if text := sb.String(); strings.TrimSpace(text) != "" {
		msg.Content = text
	} else if len(calls) == 0 {
		return apiMessage{}, false
	}
	// otherwise Content stays nil, which is the specified value for a pure tool call turn

	return msg, true
}

// toAPIUserMessages emits one role "tool" message per [completion.ToolResult] first and a single user message
// with the remaining content afterward. Images returned by a tool cannot be put into a tool message, so they are
// moved into that trailing user message.
func toAPIUserMessages(m completion.Message) ([]apiMessage, error) {
	var out []apiMessage
	var parts []apiPart

	for _, c := range m.Content {
		switch v := c.(type) {
		case completion.ToolResult:
			text, extra := toolResultContent(v)
			out = append(out, apiMessage{Role: "tool", ToolCallID: v.ToolCallID, Content: text})
			parts = append(parts, extra...)
		case completion.Text:
			if strings.TrimSpace(v.Text) == "" {
				continue
			}
			parts = append(parts, apiPart{Type: "text", Text: v.Text})
		case completion.Media:
			parts = append(parts, toAPIMediaPart(v))
		case completion.Thinking, completion.RedactedThinking:
			// not transferable
		default:
			return nil, fmt.Errorf("unsupported content type %T in user message", c)
		}
	}

	if len(parts) > 0 {
		out = append(out, apiMessage{Role: "user", Content: userContent(parts)})
	}

	return out, nil
}

// userContent returns a plain string if all parts are text, which is supported by every compatible server, and
// the list of parts otherwise.
func userContent(parts []apiPart) any {
	var sb strings.Builder
	for i, p := range parts {
		if p.Type != "text" {
			return parts
		}
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(p.Text)
	}
	return sb.String()
}

// toolResultContent flattens the result into the text of the tool message. Images are returned separately as
// parts for a following user message, all other media is inlined as text (if textual) or as a placeholder.
func toolResultContent(r completion.ToolResult) (string, []apiPart) {
	var texts []string
	var extra []apiPart

	for _, c := range r.Content {
		switch v := c.(type) {
		case completion.Text:
			texts = append(texts, v.Text)
		case completion.Media:
			part := toAPIMediaPart(v)
			if part.Type == "text" {
				texts = append(texts, part.Text)
				continue
			}
			if len(extra) == 0 {
				extra = append(extra, apiPart{Type: "text", Text: fmt.Sprintf("Attachments returned by tool call %s:", r.ToolCallID)})
			}
			extra = append(extra, part)
			texts = append(texts, fmt.Sprintf("[a file of type %s was returned and is attached to the following user message]", v.MimeType))
		}
	}

	text := strings.Join(texts, "\n")
	if r.IsError {
		text = "Error: " + text
	}

	return text, extra
}

// toAPIMediaPart maps a media block to a content part: images become image_url parts (inline data as data URL),
// textual media is inlined as text, PDFs become file parts and everything else is replaced by a textual
// placeholder, because the Chat Completions API has no generic document support.
func toAPIMediaPart(m completion.Media) apiPart {
	switch {
	case file.IsText(m.MimeType) && len(m.Source.Data) > 0:
		return apiPart{Type: "text", Text: string(m.Source.Data)}

	case isImageMime(m.MimeType) && len(m.Source.Data) > 0:
		return apiPart{Type: "image_url", ImageURL: &apiImageURL{URL: dataURL(m.MimeType, m.Source.Data)}}

	case isImageMime(m.MimeType) && m.Source.URL.IsSome():
		return apiPart{Type: "image_url", ImageURL: &apiImageURL{URL: string(m.Source.URL.Unwrap())}}

	case m.MimeType == file.PDF && len(m.Source.Data) > 0:
		return apiPart{Type: "file", File: &apiFile{FileName: "document.pdf", FileData: dataURL(m.MimeType, m.Source.Data)}}

	default:
		return apiPart{Type: "text", Text: mediaPlaceholder(m)}
	}
}

// mediaPlaceholder describes media which cannot be transferred. A file id is never sent, because this provider
// has no Files API and such ids always stem from another provider.
func mediaPlaceholder(m completion.Media) string {
	switch {
	case m.Source.URL.IsSome():
		return fmt.Sprintf("[attachment of type %s: %s]", m.MimeType, m.Source.URL.Unwrap())
	case m.Source.FileID.IsSome():
		return fmt.Sprintf("[attachment of type %s (file %s) is not available for this model]", m.MimeType, m.Source.FileID.Unwrap())
	default:
		return fmt.Sprintf("[attachment of type %s cannot be sent to this model]", m.MimeType)
	}
}

func dataURL(mime file.Type, data []byte) string {
	return "data:" + string(mime) + ";base64," + base64.StdEncoding.EncodeToString(data)
}

func isImageMime(t file.Type) bool {
	return strings.HasPrefix(strings.ToLower(string(t)), "image/")
}

// fromAPIResponse translates the first choice of a response into a [completion.Result].
func fromAPIResponse(resp apiResponse, requested model.ID) (completion.Result, error) {
	if len(resp.Choices) == 0 {
		return completion.Result{}, fmt.Errorf("openai: response contains no choices")
	}

	choice := resp.Choices[0]
	msg := choice.Message

	var content []completion.Content
	if r := msg.reasoning(); r != "" {
		content = append(content, completion.Thinking{Text: r})
	}
	if msg.Content != nil && *msg.Content != "" {
		content = append(content, completion.Text{Text: *msg.Content})
	}
	refused := msg.Refusal != nil && *msg.Refusal != ""
	if refused {
		content = append(content, completion.Text{Text: *msg.Refusal})
	}
	for _, tc := range msg.ToolCalls {
		content = append(content, fromAPIToolCall(tc.ID, tc.Function.Name, tc.Function.Arguments))
	}

	stop := fromAPIFinishReason(choice.FinishReason, len(msg.ToolCalls) > 0)
	if refused && stop == completion.StopEndTurn {
		stop = completion.StopRefusal
	}

	m := model.ID(resp.Model)
	if m == "" {
		m = requested
	}

	return completion.Result{
		Message: completion.Message{
			Role:    completion.Assistant,
			Content: content,
		},
		StopReason: stop,
		Usage:      fromAPIUsage(resp.Usage),
		Model:      m,
	}, nil
}

// fromAPIToolCall builds a tool call, generating an id if the server sent none (some compatible servers do
// not), because the id is required to pair the result with the call.
func fromAPIToolCall(id, name, args string) completion.ToolCall {
	if id == "" {
		id = newToolCallID()
	}

	raw := json.RawMessage(strings.TrimSpace(args))
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}

	return completion.ToolCall{ID: id, Name: name, Arguments: raw}
}

func newToolCallID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return "call_" + hex.EncodeToString(b[:])
}

// fromAPIUsage maps the token usage. Following the Anthropic semantics of [completion.Usage], InputTokens
// excludes the cached prompt tokens, which OpenAI includes in prompt_tokens and reports separately.
func fromAPIUsage(u *apiUsage) completion.Usage {
	if u == nil {
		return completion.Usage{}
	}

	cached := 0
	if u.PromptTokensDetails != nil {
		cached = u.PromptTokensDetails.CachedTokens
	}

	return completion.Usage{
		InputTokens:     max(u.PromptTokens-cached, 0),
		OutputTokens:    u.CompletionTokens,
		CacheReadTokens: cached,
	}
}

// fromAPIFinishReason maps finish_reason to a [completion.StopReason]. Some compatible servers report "stop"
// although the model called tools, so the presence of tool calls wins over the reported reason.
func fromAPIFinishReason(reason string, hasToolCalls bool) completion.StopReason {
	switch reason {
	case "length":
		return completion.StopMaxTokens
	case "content_filter":
		return completion.StopRefusal
	case "tool_calls", "function_call":
		return completion.StopToolUse
	}

	if hasToolCalls {
		return completion.StopToolUse
	}

	switch reason {
	case "stop", "", "eos", "end_turn":
		return completion.StopEndTurn
	case "stop_sequence":
		return completion.StopStopSequence
	default:
		return completion.StopReason(reason)
	}
}
