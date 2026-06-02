// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"encoding/json"
	"fmt"
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
)

var _ completion.Completions = (*anthropicCompletions)(nil)

type anthropicCompletions struct {
	parent *anthropicProvider
}

func (c *anthropicCompletions) client() *Client {
	return c.parent.client()
}

func (c *anthropicCompletions) Models(subject auth.Subject) iter.Seq2[model.Model, error] {
	return c.parent.listModels(subject)
}

func (c *anthropicCompletions) Complete(subject auth.Subject, opts completion.Options) (completion.Result, error) {

	if len(opts.Messages) == 0 {
		return completion.Result{}, fmt.Errorf("messages must not be empty")
	}

	req, err := c.parent.buildRequest(opts)
	if err != nil {
		return completion.Result{}, err
	}

	resp, err := c.client().CreateMessage(req)
	if err != nil {
		return completion.Result{}, err
	}

	return fromAPIResponse(resp), nil
}

func (c *anthropicCompletions) Stream(subject auth.Subject, opts completion.Options) iter.Seq2[completion.Delta, error] {
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

		// streaming state
		var (
			usage      completion.Usage
			stopReason completion.StopReason
			// per content block index bookkeeping for tool_use accumulation
			toolBlocks = map[int]*toolAccu{}
			aborted    bool
		)

		emit := func(d completion.Delta) bool {
			if !yield(d, nil) {
				aborted = true
				return false
			}
			return true
		}

		err = c.client().CreateMessageStream(req, func(event string, data []byte) error {
			if aborted {
				return errStopStreaming
			}

			switch event {
			case "message_start":
				var ev struct {
					Message struct {
						Usage apiUsage `json:"usage"`
					} `json:"message"`
				}
				if err := json.Unmarshal(data, &ev); err != nil {
					return err
				}
				usage.InputTokens = ev.Message.Usage.InputTokens
				usage.CacheReadTokens = ev.Message.Usage.CacheReadInputTokens
				usage.CacheWriteTokens = ev.Message.Usage.CacheCreationInputTokens

			case "content_block_start":
				var ev struct {
					Index        int        `json:"index"`
					ContentBlock apiContent `json:"content_block"`
				}
				if err := json.Unmarshal(data, &ev); err != nil {
					return err
				}
				if ev.ContentBlock.Type == "tool_use" {
					toolBlocks[ev.Index] = &toolAccu{id: ev.ContentBlock.ID, name: ev.ContentBlock.Name}
				}

			case "content_block_delta":
				var ev struct {
					Index int `json:"index"`
					Delta struct {
						Type        string `json:"type"`
						Text        string `json:"text"`
						PartialJSON string `json:"partial_json"`
						Thinking    string `json:"thinking"`
					} `json:"delta"`
				}
				if err := json.Unmarshal(data, &ev); err != nil {
					return err
				}
				switch ev.Delta.Type {
				case "text_delta":
					if ev.Delta.Text != "" {
						if !emit(completion.Delta{TextDelta: ev.Delta.Text}) {
							return errStopStreaming
						}
					}
				case "input_json_delta":
					if acc := toolBlocks[ev.Index]; acc != nil {
						acc.json += ev.Delta.PartialJSON
					}
				case "thinking_delta":
					// thinking is streamed but currently surfaced only via the final non-streaming
					// representation; ignore incremental thinking here.
				}

			case "content_block_stop":
				var ev struct {
					Index int `json:"index"`
				}
				if err := json.Unmarshal(data, &ev); err != nil {
					return err
				}
				if acc := toolBlocks[ev.Index]; acc != nil {
					args := json.RawMessage(acc.json)
					if len(args) == 0 {
						args = json.RawMessage("{}")
					}
					call := completion.ToolCall{ID: acc.id, Name: acc.name, Arguments: args}
					if !emit(completion.Delta{ToolCall: option.Some(call)}) {
						return errStopStreaming
					}
					delete(toolBlocks, ev.Index)
				}

			case "message_delta":
				var ev struct {
					Delta struct {
						StopReason string `json:"stop_reason"`
					} `json:"delta"`
					Usage struct {
						OutputTokens int `json:"output_tokens"`
					} `json:"usage"`
				}
				if err := json.Unmarshal(data, &ev); err != nil {
					return err
				}
				if ev.Delta.StopReason != "" {
					stopReason = fromAPIStopReason(ev.Delta.StopReason)
				}
				if ev.Usage.OutputTokens > 0 {
					usage.OutputTokens = ev.Usage.OutputTokens
				}

			case "error":
				return fmt.Errorf("anthropic stream error: %s", string(data))
			}

			return nil
		})

		if aborted {
			return
		}

		if err != nil {
			yield(completion.Delta{}, err)
			return
		}

		emit(completion.Delta{
			Done:       true,
			StopReason: stopReason,
			Usage:      option.Some(usage),
		})
	}
}

type toolAccu struct {
	id   string
	name string
	json string
}

// errStopStreaming is an internal sentinel used to unwind the SSE callback once the consumer stopped
// iterating. It is swallowed by the [anthropicCompletions.Stream] implementation.
var errStopStreaming = fmt.Errorf("stop streaming")

