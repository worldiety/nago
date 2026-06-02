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
		System:        opts.System,
		StopSequences: opts.StopSequences,
		Metadata:      opts.Metadata,
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

	return req, nil
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
		nested, err := toAPIContents(v.Content)
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
	case "end_turn", "pause_turn":
		return completion.StopEndTurn
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

