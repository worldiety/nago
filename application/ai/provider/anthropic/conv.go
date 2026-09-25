// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"encoding/base64"
	"fmt"
	"strings"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/model"
)

// buildRequest translates the stateless [completion.Options] into the Anthropic wire request.
func (p *anthropicProvider) buildRequest(opts completion.Options) (apiRequest, error) {
	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = p.cfg.MaxTokens
	}
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	req := apiRequest{
		Model:         string(opts.Model),
		MaxTokens:     maxTokens,
		StopSequences: opts.StopSequences,
		Metadata:      opts.Metadata,
	}

	if opts.System != "" {
		req.System = []apiContent{{Type: "text", Text: opts.System}}
	}

	if opts.Temperature.IsSome() {
		v := opts.Temperature.Unwrap()
		req.Temperature = &v
	}

	if opts.TopP.IsSome() {
		v := opts.TopP.Unwrap()
		req.TopP = &v
	}

	for _, m := range opts.Messages {
		am, err := toAPIMessage(m)
		if err != nil {
			return apiRequest{}, err
		}
		// A message without content is rejected by the API. Dropping it is safe, because Anthropic merges
		// consecutive turns of the same role anyway.
		if len(am.Content) == 0 {
			continue
		}
		req.Messages = append(req.Messages, am)
	}

	for _, t := range opts.Tools {
		req.Tools = append(req.Tools, apiTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Schema,
		})
	}

	if tc, ok := toAPIToolChoice(opts.ToolChoice); ok {
		req.ToolChoice = &tc
	}

	req.Thinking = p.thinking(opts, &req)

	if !p.cfg.DisablePromptCache {
		p.applyPromptCache(&req)
	}

	return req, nil
}

// minThinkingBudget is the smallest budget_tokens value accepted by the API.
const minThinkingBudget = 1024

// thinking resolves the thinking configuration for the request. For [completion.ThinkingAuto] adaptive
// thinking is used, unless the request uses options the API does not allow together with thinking (custom
// temperature/top_p or a forced tool choice) or the model is known to reject adaptive thinking. A budgeted
// request raises max_tokens if needed, because budget_tokens must be smaller than max_tokens.
func (p *anthropicProvider) thinking(opts completion.Options, req *apiRequest) *apiThinking {
	switch opts.Thinking {
	case completion.ThinkingOff:
		return nil
	case completion.ThinkingAdaptive:
		return &apiThinking{Type: "adaptive"}
	case completion.ThinkingBudgeted:
		budget := max(opts.ThinkingBudget, minThinkingBudget)
		if req.MaxTokens <= budget {
			req.MaxTokens = budget + defaultMaxTokens
		}
		return &apiThinking{Type: "enabled", BudgetTokens: budget}
	default: // auto
		if req.Temperature != nil || req.TopP != nil {
			return nil
		}
		if req.ToolChoice != nil && (req.ToolChoice.Type == "any" || req.ToolChoice.Type == "tool") {
			return nil
		}
		if _, rejected := p.noAdaptiveThinking.Load(req.Model); rejected {
			return nil
		}
		return &apiThinking{Type: "adaptive"}
	}
}

// applyPromptCache places Anthropic prompt-cache breakpoints on the stable request prefix. Anthropic hashes
// the prefix (tools -> system -> messages) up to and including each marked block, so only the most stable
// blocks are marked. At most 4 breakpoints are allowed; we use up to three, in order of stability:
//
//  1. the last tool definition (tools never change within a run),
//  2. the system prompt,
//  3. the last content block of the second-to-last message, i.e. the frozen conversation history just before
//     the newest (uncached) turn. In the agentic loop this breakpoint advances automatically as the history
//     grows, so every turn reads the previous prefix from cache.
func (p *anthropicProvider) applyPromptCache(req *apiRequest) {
	cc := p.cacheControl()

	// 1. last tool
	if n := len(req.Tools); n > 0 {
		req.Tools[n-1].CacheControl = cc
	}

	// 2. system prompt
	if n := len(req.System); n > 0 {
		req.System[n-1].CacheControl = cc
	}

	// 3. frozen history boundary: the last cacheable block of the second-to-last message. With only a single
	// message there is no stable prefix to cache yet, so we skip it. Thinking blocks must not carry
	// cache_control (the API rejects it), so we walk backwards to the nearest cacheable block, possibly into
	// earlier messages. If none exists, the breakpoint is omitted.
	for i := len(req.Messages) - 2; i >= 0; i-- {
		content := req.Messages[i].Content
		for j := len(content) - 1; j >= 0; j-- {
			if isCacheableBlock(content[j].Type) {
				content[j].CacheControl = cc
				return
			}
		}
	}
}

// isCacheableBlock reports whether the Anthropic API accepts a cache_control marker on a block of the given
// type. Thinking and redacted thinking blocks are rejected with "Extra inputs are not permitted".
func isCacheableBlock(typ string) bool {
	switch typ {
	case "thinking", "redacted_thinking":
		return false
	default:
		return true
	}
}

// cacheControl builds the ephemeral cache marker honoring the configured TTL.
func (p *anthropicProvider) cacheControl() *apiCacheControl {
	cc := &apiCacheControl{Type: "ephemeral"}
	if p.cfg.PromptCacheTTL == "1h" {
		cc.TTL = "1h"
	}
	return cc
}

func toAPIToolChoice(tc completion.ToolChoice) (apiToolChoice, bool) {
	switch {
	case tc.Name != "":
		return apiToolChoice{Type: "tool", Name: tc.Name}, true
	case tc.Mode == "any":
		return apiToolChoice{Type: "any"}, true
	case tc.Mode == "none":
		return apiToolChoice{Type: "none"}, true
	case tc.Mode == "auto":
		return apiToolChoice{Type: "auto"}, true
	default:
		return apiToolChoice{}, false
	}
}

func toAPIMessage(m completion.Message) (apiMessage, error) {
	content, err := toAPIContents(m.Content)
	if err != nil {
		return apiMessage{}, err
	}

	return apiMessage{
		Role:    string(m.Role),
		Content: content,
	}, nil
}

func toAPIContents(in []completion.Content) ([]apiContent, error) {
	out := make([]apiContent, 0, len(in))
	for _, c := range in {
		// The API rejects empty text blocks ("text content blocks must be non-empty").
		if t, ok := c.(completion.Text); ok && strings.TrimSpace(t.Text) == "" {
			continue
		}

		ac, err := toAPIContent(c)
		if err != nil {
			return nil, err
		}
		out = append(out, ac)
	}
	return out, nil
}

func toAPIContent(c completion.Content) (apiContent, error) {
	switch v := c.(type) {
	case completion.Text:
		return apiContent{Type: "text", Text: v.Text}, nil

	case completion.Thinking:
		return apiContent{Type: "thinking", Thinking: v.Text, Signature: v.Signature}, nil

	case completion.RedactedThinking:
		return apiContent{Type: "redacted_thinking", Data: v.Data}, nil

	case completion.Media:
		src, err := toAPISource(v.MimeType, v.Source)
		if err != nil {
			return apiContent{}, err
		}

		blockType := "document"
		if isImageMime(v.MimeType) {
			blockType = "image"
		}

		return apiContent{Type: blockType, Source: &src}, nil

	case completion.ToolCall:
		input := v.Arguments
		if len(input) == 0 {
			input = []byte("{}")
		}
		return apiContent{Type: "tool_use", ID: v.ID, Name: v.Name, Input: input}, nil

	case completion.ToolResult:
		// Anthropic restricts what may appear inside a tool_result: only text and image blocks are allowed,
		// and image blocks there accept only base64/url sources — never a file id. Normalize the content
		// first so inline media that cannot be embedded becomes descriptive text instead of a block the API
		// rejects.
		nested, err := toAPIContents(normalizeToolResultContent(v.Content))
		if err != nil {
			return apiContent{}, err
		}
		return apiContent{
			Type:      "tool_result",
			ToolUseID: v.ToolCallID,
			Content:   nested,
			IsError:   v.IsError,
		}, nil

	default:
		return apiContent{}, fmt.Errorf("unsupported content type %T", c)
	}
}

// normalizeToolResultContent adapts tool-result content blocks to the subset Anthropic accepts inside a
// tool_result. Only text and image blocks are allowed there, and image blocks accept only base64/url sources.
//
// Consequences:
//   - [completion.Text] passes through unchanged.
//   - A [completion.Media] with inline base64 image data passes through (base64 image sources are allowed);
//     any other inline media (non-image, or an image sourced by url/file id) is replaced by a text reference,
//     because embedding it here would either be rejected by the API or bloat every follow-up turn.
func normalizeToolResultContent(in []completion.Content) []completion.Content {
	out := make([]completion.Content, 0, len(in))
	for _, c := range in {
		switch v := c.(type) {
		case completion.Media:
			// Only inline base64 image data is permitted as an image block inside a tool_result.
			if isImageMime(v.MimeType) && len(v.Source.Data) > 0 {
				out = append(out, v)
				continue
			}
			out = append(out, completion.Text{Text: fmt.Sprintf(
				"[a file of type %s was produced but cannot be embedded in a tool result]", v.MimeType)})

		default:
			out = append(out, c)
		}
	}
	return out
}

func toAPISource(mime file.Type, src completion.Source) (apiSource, error) {
	switch {
	case len(src.Data) > 0:
		return apiSource{
			Type:      "base64",
			MediaType: string(mime),
			Data:      base64.StdEncoding.EncodeToString(src.Data),
		}, nil
	case src.URL.IsSome():
		return apiSource{Type: "url", URL: string(src.URL.Unwrap())}, nil
	case src.FileID.IsSome():
		return apiSource{Type: "file", FileID: string(src.FileID.Unwrap())}, nil
	default:
		return apiSource{}, fmt.Errorf("media content has no source data, url or file id")
	}
}

func isImageMime(t file.Type) bool {
	switch t {
	case file.PNG, file.JPEG, file.GIF:
		return true
	default:
		return false
	}
}

// fromAPIResponse translates an Anthropic Messages response into a [completion.Result].
func fromAPIResponse(resp apiResponse) completion.Result {
	return completion.Result{
		Message: completion.Message{
			Role:    completion.Assistant,
			Content: fromAPIContents(resp.Content),
		},
		StopReason: fromAPIStopReason(resp.StopReason),
		Usage:      fromAPIUsage(resp.Usage),
		Model:      model.ID(resp.Model),
	}
}

func fromAPIContents(in []apiContent) []completion.Content {
	out := make([]completion.Content, 0, len(in))
	for _, c := range in {
		switch c.Type {
		case "text":
			out = append(out, completion.Text{Text: c.Text})
		case "thinking":
			out = append(out, completion.Thinking{Text: c.Thinking, Signature: c.Signature})
		case "redacted_thinking":
			out = append(out, completion.RedactedThinking{Data: c.Data})
		case "tool_use":
			out = append(out, completion.ToolCall{ID: c.ID, Name: c.Name, Arguments: c.Input})
		}
	}
	return out
}

func fromAPIUsage(u apiUsage) completion.Usage {
	return completion.Usage{
		InputTokens:      u.InputTokens,
		OutputTokens:     u.OutputTokens,
		CacheReadTokens:  u.CacheReadInputTokens,
		CacheWriteTokens: u.CacheCreationInputTokens,
	}
}

func fromAPIStopReason(reason string) completion.StopReason {
	switch reason {
	case "end_turn":
		return completion.StopEndTurn
	case "pause_turn":
		return completion.StopPauseTurn
	case "max_tokens":
		return completion.StopMaxTokens
	case "stop_sequence":
		return completion.StopStopSequence
	case "tool_use":
		return completion.StopToolUse
	case "refusal":
		return completion.StopRefusal
	default:
		return completion.StopReason(reason)
	}
}
