// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"fmt"
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
)

var _ completion.Completions = (*gollamaCompletions)(nil)

type gollamaCompletions struct {
	parent *gollamaProvider
}

func (c *gollamaCompletions) Models(subject auth.Subject) iter.Seq2[model.Model, error] {
	return listModels()
}

func (c *gollamaCompletions) Complete(subject auth.Subject, opts completion.Options) (completion.Result, error) {
	lm, nCtx, err := c.prepare(opts)
	if err != nil {
		return completion.Result{}, err
	}

	prompt, err := lm.adapter.renderPrompt(opts)
	if err != nil {
		return completion.Result{}, fmt.Errorf("render prompt: %w", err)
	}

	out, err := c.parent.eng.generate(lm, prompt, lm.adapter.stopStrings(), lm.adapter.toolMarkers(), opts.MaxTokens, nCtx, nil)
	if err != nil {
		return completion.Result{}, err
	}

	contents, parsed := lm.adapter.parse(out.text)

	return completion.Result{
		Message: completion.Message{
			Role:    completion.Assistant,
			Content: contents,
		},
		StopReason: combineStopReason(out.stop, parsed),
		Usage: completion.Usage{
			InputTokens:  out.promptTokens,
			OutputTokens: out.outputTokens,
		},
		Model: opts.Model,
	}, nil
}

func (c *gollamaCompletions) Stream(subject auth.Subject, opts completion.Options) iter.Seq2[completion.Delta, error] {
	return func(yield func(completion.Delta, error) bool) {
		lm, nCtx, err := c.prepare(opts)
		if err != nil {
			yield(completion.Delta{}, err)
			return
		}

		prompt, err := lm.adapter.renderPrompt(opts)
		if err != nil {
			yield(completion.Delta{}, fmt.Errorf("render prompt: %w", err))
			return
		}

		var aborted bool
		emit := func(s string) bool {
			if !yield(completion.Delta{TextDelta: s}, nil) {
				aborted = true
				return false
			}
			return true
		}

		out, err := c.parent.eng.generate(lm, prompt, lm.adapter.stopStrings(), lm.adapter.toolMarkers(), opts.MaxTokens, nCtx, emit)
		if aborted {
			return
		}
		if err != nil {
			yield(completion.Delta{}, err)
			return
		}

		contents, parsed := lm.adapter.parse(out.text)

		var hasTool bool
		for _, ct := range contents {
			if tc, ok := ct.(completion.ToolCall); ok {
				hasTool = true
				if !yield(completion.Delta{ToolCall: option.Some(tc)}, nil) {
					return
				}
			}
		}

		// If text was held back by tool-call masking but no tool call materialised, surface the remainder now
		// so no assistant text is lost.
		if !hasTool && out.emitted < len(out.text) {
			if !yield(completion.Delta{TextDelta: out.text[out.emitted:]}, nil) {
				return
			}
		}

		yield(completion.Delta{
			Done:       true,
			StopReason: combineStopReason(out.stop, parsed),
			Usage: option.Some(completion.Usage{
				InputTokens:  out.promptTokens,
				OutputTokens: out.outputTokens,
			}),
		}, nil)
	}
}

// prepare validates the request, resolves and loads the model and computes the effective context window.
func (c *gollamaCompletions) prepare(opts completion.Options) (*loadedModel, int, error) {
	if len(opts.Messages) == 0 {
		return nil, 0, fmt.Errorf("messages must not be empty")
	}

	entry, ok := lookupCatalog(opts.Model)
	if !ok {
		return nil, 0, fmt.Errorf("unknown model %q", opts.Model)
	}

	lm, err := c.parent.eng.load(entry)
	if err != nil {
		return nil, 0, err
	}

	return lm, c.parent.eng.effectiveCtx(entry, lm.meta), nil
}

// combineStopReason reconciles the engine's raw stop reason with the adapter's parse result: a parsed tool
// call always wins, because the agentic loop keys off [completion.StopToolUse].
func combineStopReason(raw, parsed completion.StopReason) completion.StopReason {
	if parsed == completion.StopToolUse {
		return completion.StopToolUse
	}
	return raw
}
