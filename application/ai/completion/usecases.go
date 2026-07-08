// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package completion is a PROPOSAL for a stateless message API for the nago ai abstraction.
//
// In contrast to [conversation.Conversation], which models a server-side stored chat (Mistral/OpenAI
// "conversations", Anthropic does not store anything), a completion is fully stateless: the caller
// always submits the entire history and receives a single assistant turn back. This maps 1:1 to:
//
//   - Anthropic "Messages" API   -> POST /v1/messages
//   - OpenAI "Chat Completions"  -> POST /v1/chat/completions
//   - OpenAI "Responses" API     -> POST /v1/responses
//
// The design deliberately follows the Anthropic content-block model, because it is the most general:
// system prompt is a dedicated field, tool results are carried as content blocks inside a user turn.
// An OpenAI provider can trivially translate this (system block -> system message, tool_result block ->
// role:"tool" message, tool_use block -> assistant.tool_calls).
package completion

import (
	"encoding/json"
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

// Role of a message within the stateless history. We intentionally reuse only the two transport roles.
// A "system" prompt is modelled as [Options.System] (Anthropic style) and tool results are modelled as
// [ToolResult] content blocks inside a User message, so no extra roles are required.
type Role string

const (
	User      Role = "user"
	Assistant Role = "assistant"
)

// Options is the stateless request. It contains the complete conversation history on every call.
type Options struct {
	// Model to run the completion against. Required.
	Model model.ID

	// System is the system / developer instruction. Maps to Anthropic "system" and to an OpenAI
	// system/developer message. Optional.
	System string

	// Messages is the full, ordered history. Must not be empty and should usually end with a User message
	// (or a User message carrying [ToolResult] blocks when continuing after a tool call).
	Messages []Message

	// Tools advertises the function tools the model may call. Optional.
	Tools []ToolDef

	// ToolChoice controls whether/which tool must be used. Optional (zero value = auto).
	ToolChoice ToolChoice

	// MaxTokens caps the generated output tokens. Anthropic requires this; OpenAI treats it as optional.
	MaxTokens int

	// Temperature in [0..1]. Optional.
	Temperature option.Opt[float64]

	// TopP nucleus sampling. Optional.
	TopP option.Opt[float64]

	// StopSequences are custom strings that stop generation. Optional.
	StopSequences []string

	// Metadata is opaque provider metadata (e.g. Anthropic metadata.user_id). Optional.
	Metadata map[string]string
}

// Message is one turn in the stateless history.
type Message struct {
	Role    Role      `json:"role"`
	Content []Content `json:"content"`
}

// Content is the union of all supported content blocks. Use the concrete types below.
type Content interface {
	isContent()
}

// Text is a plain text block (Anthropic "text", OpenAI "text"/string content).
type Text struct {
	Text string `json:"text"`
}

func (Text) isContent() {}

// Source describes how binary content (image/document) is provided. Exactly one of the fields is set.
// This mirrors Anthropic's source.{base64|url|file} and OpenAI's image_url/file/file_id.
type Source struct {
	// Data is inline binary content; the provider base64-encodes it.
	Data []byte `json:"data,omitzero"`
	// URL references externally hosted content.
	URL option.Opt[core.URI] `json:"url,omitzero"`
	// FileID references a file previously uploaded via provider.Files() (Anthropic Files API / OpenAI files).
	FileID option.Opt[file.ID] `json:"fileId,omitzero"`
}

// Media is an image or document content block. We keep the existing [file.Type] mime enumeration.
type Media struct {
	MimeType file.Type `json:"mimeType"`
	Source   Source    `json:"source"`
}

func (Media) isContent() {}

// FileRef references a file managed by a provider's Files capability (uploaded via [provider.Files.Put]) so
// it can be sent to the model by its provider-native file id instead of inlining the bytes. This keeps the
// request small and efficient: the provider fetches the file once, while the reference stays O(1) in size on
// every follow-up turn of an agentic loop (where the full history is re-sent each turn).
//
// A FileRef is bound to the provider of the active completion; it deliberately does not carry a provider id.
// Obtain one from the result of an upload rather than constructing it ad hoc, so it always points at a file
// the active provider can actually resolve.
type FileRef struct {
	// File is the provider-native file identifier returned by the upload.
	File file.ID `json:"file"`
	// MimeType lets the provider pick the correct block type (image vs. document) without re-fetching.
	MimeType file.Type `json:"mimeType"`
}

func (FileRef) isContent() {}

// ToolCall is emitted by the assistant when it wants to call a tool (Anthropic "tool_use",
// OpenAI assistant.tool_calls[]).
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (ToolCall) isContent() {}

// ToolResult carries the result of a tool execution back to the model. In a request it lives inside a
// User message (Anthropic "tool_result"); an OpenAI provider maps it to a role:"tool" message.
type ToolResult struct {
	ToolCallID string    `json:"toolCallId"`
	Content    []Content `json:"content"`
	IsError    bool      `json:"isError,omitzero"`
}

func (ToolResult) isContent() {}

// Thinking is an extended-reasoning block (Anthropic "thinking"). Usually only present in responses.
type Thinking struct {
	Text      string `json:"text"`
	Signature string `json:"signature,omitzero"`
}

func (Thinking) isContent() {}

// ToolDef advertises a callable function tool. Schema is a JSON-Schema object describing the input.
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitzero"`
	Schema      json.RawMessage `json:"schema"`
}

// ToolChoice controls tool usage. Zero value means "auto".
type ToolChoice struct {
	// Mode auto|any|none. Empty == auto.
	Mode string `json:"mode,omitzero"`
	// Name forces a specific tool when set (Anthropic tool_choice.name / OpenAI named tool_choice).
	Name string `json:"name,omitzero"`
}

// StopReason explains why generation stopped.
type StopReason string

const (
	StopEndTurn      StopReason = "end_turn"
	StopMaxTokens    StopReason = "max_tokens"
	StopStopSequence StopReason = "stop_sequence"
	StopToolUse      StopReason = "tool_use"
	StopRefusal      StopReason = "refusal"
)

// Usage reports token accounting. Cache fields are optional and may be zero for providers without prompt
// caching.
type Usage struct {
	InputTokens      int `json:"inputTokens"`
	OutputTokens     int `json:"outputTokens"`
	CacheReadTokens  int `json:"cacheReadTokens,omitzero"`
	CacheWriteTokens int `json:"cacheWriteTokens,omitzero"`
}

// Result is the single assistant turn returned by a completion.
type Result struct {
	// Message is the generated assistant message. Its Content may contain Text, Thinking and/or ToolCall
	// blocks. When StopReason == StopToolUse, feed the ToolCall results back as a follow-up User message
	// containing ToolResult blocks.
	Message Message `json:"message"`

	StopReason StopReason `json:"stopReason"`
	Usage      Usage      `json:"usage"`
	Model      model.ID   `json:"model"`
}

// Delta is a single streaming chunk. Either TextDelta or a (partial) ToolCall is set; the final chunk
// carries Done == true together with StopReason and Usage.
type Delta struct {
	TextDelta  string         `json:"textDelta,omitzero"`
	ToolCall   option.Opt[ToolCall] `json:"toolCall,omitzero"`
	Done       bool           `json:"done,omitzero"`
	StopReason StopReason     `json:"stopReason,omitzero"`
	Usage      option.Opt[Usage]    `json:"usage,omitzero"`
}

// Completions is the stateless capability surface a Provider may expose.
type Completions interface {
	// Models lists the models usable for stateless completions.
	Models(subject auth.Subject) iter.Seq2[model.Model, error]

	// Complete runs a single, blocking, stateless turn over the supplied history.
	// Defined errors:
	//   - provider.TooManyRequests if the rate limiter kicked in.
	Complete(subject auth.Subject, opts Options) (Result, error)

	// Stream runs the same request but yields incremental [Delta] chunks. Providers that cannot stream may
	// return option.None for [Streaming].
	Stream(subject auth.Subject, opts Options) iter.Seq2[Delta, error]
}

