// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package openai

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"strings"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
)

var _ completion.Completions = (*openaiCompletions)(nil)

type openaiCompletions struct {
	parent *openaiProvider
}

func (c *openaiCompletions) client() *Client {
	return c.parent.client()
}

func (c *openaiCompletions) Models(subject auth.Subject) iter.Seq2[model.Model, error] {
	return c.parent.listModels(subject)
}

func (c *openaiCompletions) Complete(ctx context.Context, subject auth.Subject, opts completion.Options) (completion.Result, error) {
	if len(opts.Messages) == 0 {
		return completion.Result{}, fmt.Errorf("messages must not be empty")
	}

	req, err := c.parent.buildRequest(opts)
	if err != nil {
		return completion.Result{}, err
	}

	resp, err := c.client().ChatCompletion(ctx, req)
	for attempt := 0; err != nil && attempt < maxAdjustments && c.parent.adjustRequest(&req, err); attempt++ {
		resp, err = c.client().ChatCompletion(ctx, req)
	}
	if err != nil {
		return completion.Result{}, err
	}

	return fromAPIResponse(resp, opts.Model)
}

// maxAdjustments caps how often a rejected request is adjusted and repeated, see
// [openaiProvider.adjustRequest].
const maxAdjustments = 2

// adjustRequest adapts req to a server which rejected an optional or version dependent field and reports
// whether the request should be repeated. The decision is remembered for subsequent requests. It covers:
//   - max_tokens vs. max_completion_tokens: OpenAI reasoning models reject the former, older compatible servers
//     do not know the latter,
//   - stream_options, which not every compatible server supports.
func (p *openaiProvider) adjustRequest(req *apiRequest, err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode < 400 || apiErr.StatusCode >= 500 || apiErr.StatusCode == http.StatusTooManyRequests {
		return false
	}

	msg := strings.ToLower(apiErr.Message)

	if req.StreamOptions != nil && strings.Contains(msg, "stream_options") {
		req.StreamOptions = nil
		p.noStreamUsage.Store(true)
		return true
	}

	switch {
	case req.MaxTokens > 0 && strings.Contains(msg, "max_completion_tokens"):
		// e.g. "Unsupported parameter: 'max_tokens' is not supported with this model. Use
		// 'max_completion_tokens' instead."
		req.MaxCompletionTokens, req.MaxTokens = req.MaxTokens, 0
		p.useMaxCompletionTokens.Store(true)
		return true
	case req.MaxCompletionTokens > 0 && strings.Contains(msg, "max_completion_tokens"):
		// the server does not know the newer field
		req.MaxTokens, req.MaxCompletionTokens = req.MaxCompletionTokens, 0
		p.useMaxCompletionTokens.Store(false)
		return true
	}

	return false
}

func (c *openaiCompletions) Stream(ctx context.Context, subject auth.Subject, opts completion.Options) iter.Seq2[completion.Delta, error] {
	return func(yield func(completion.Delta, error) bool) {
		if len(opts.Messages) == 0 {
			yield(completion.Delta{}, fmt.Errorf("messages must not be empty"))
			return
		}

		req, err := c.parent.buildRequest(opts)
		if err != nil {
			yield(completion.Delta{}, err)
			return
		}

		if !c.parent.noStreamUsage.Load() {
			req.StreamOptions = &apiStreamOptions{IncludeUsage: true}
		}

		var (
			usage      completion.Usage
			stopReason completion.StopReason
			tools      toolAccus
			started    bool
			aborted    bool
		)

		emit := func(d completion.Delta) bool {
			if !yield(d, nil) {
				aborted = true
				return false
			}
			return true
		}

		// flushTools emits the accumulated tool calls. They are only complete once the choice finished,
		// because the arguments are streamed in fragments.
		flushTools := func() bool {
			for _, acc := range tools.drain() {
				call := fromAPIToolCall(acc.id, acc.name, acc.args.String())
				if !emit(completion.Delta{ToolCall: option.Some(call)}) {
					return false
				}
			}
			return true
		}

		onChunk := func(chunk apiChunk) error {
			started = true
			if aborted {
				return errStopStreaming
			}

			if chunk.Usage != nil {
				usage = fromAPIUsage(chunk.Usage)
			}

			for _, choice := range chunk.Choices {
				if choice.Index != 0 {
					continue // we never request n > 1
				}

				if d := choice.Delta.Content; d != nil && *d != "" {
					if !emit(completion.Delta{TextDelta: *d}) {
						return errStopStreaming
					}
				}

				if d := choice.Delta.Refusal; d != nil && *d != "" {
					stopReason = completion.StopRefusal
					if !emit(completion.Delta{TextDelta: *d}) {
						return errStopStreaming
					}
				}

				// reasoning deltas (reasoning_content/reasoning) are not surfaced, because a Delta cannot carry
				// thinking; this matches the anthropic provider.

				for _, tc := range choice.Delta.ToolCalls {
					tools.add(tc)
				}

				if choice.FinishReason != nil && *choice.FinishReason != "" {
					reason := fromAPIFinishReason(*choice.FinishReason, tools.len() > 0)
					if stopReason != completion.StopRefusal || reason != completion.StopEndTurn {
						stopReason = reason
					}
					if !flushTools() {
						return errStopStreaming
					}
				}
			}

			return nil
		}

		err = c.client().ChatCompletionStream(ctx, req, onChunk)
		for attempt := 0; err != nil && !started && attempt < maxAdjustments && c.parent.adjustRequest(&req, err); attempt++ {
			err = c.client().ChatCompletionStream(ctx, req, onChunk)
		}

		if aborted {
			return
		}

		if err != nil {
			yield(completion.Delta{}, err)
			return
		}

		// A server which ended the stream without a finish_reason still may have sent tool calls.
		if tools.len() > 0 {
			if stopReason == "" || stopReason == completion.StopEndTurn {
				stopReason = completion.StopToolUse
			}
			if !flushTools() {
				return
			}
		}

		if stopReason == "" {
			stopReason = completion.StopEndTurn
		}

		emit(completion.Delta{
			Done:       true,
			StopReason: stopReason,
			Usage:      option.Some(usage),
		})
	}
}

// toolAccu collects the fragments of a single streamed tool call.
type toolAccu struct {
	id   string
	name string
	args strings.Builder
}

// toolAccus collects the streamed tool calls in the order of their first appearance. The spec identifies a call
// by its index and sends id and name only with the first fragment. Some compatible servers, however, send every
// call complete in a single fragment and reuse index 0 (or omit it), so a fragment carrying a new id always
// starts a new call.
type toolAccus struct {
	list    []*toolAccu
	byIndex map[int]*toolAccu
}

func (t *toolAccus) add(tc apiToolCall) {
	if t.byIndex == nil {
		t.byIndex = map[int]*toolAccu{}
	}

	continuation := tc.ID == "" && tc.Function.Name == ""

	var acc *toolAccu
	if tc.Index != nil {
		acc = t.byIndex[*tc.Index]
		if acc != nil && tc.ID != "" && acc.id != "" && tc.ID != acc.id {
			acc = nil // a new call reusing the index
		}
	} else if continuation && len(t.list) > 0 {
		acc = t.list[len(t.list)-1]
	}

	if acc == nil {
		acc = &toolAccu{}
		t.list = append(t.list, acc)
		if tc.Index != nil {
			t.byIndex[*tc.Index] = acc
		}
	}

	if tc.ID != "" {
		acc.id = tc.ID
	}
	if tc.Function.Name != "" {
		acc.name = tc.Function.Name
	}
	acc.args.WriteString(tc.Function.Arguments)
}

func (t *toolAccus) len() int {
	return len(t.list)
}

func (t *toolAccus) drain() []*toolAccu {
	list := t.list
	t.list = nil
	t.byIndex = nil
	return list
}

// errStopStreaming is an internal sentinel used to unwind the SSE callback once the consumer stopped
// iterating. It is swallowed by the [openaiCompletions.Stream] implementation.
var errStopStreaming = errors.New("stop streaming")
