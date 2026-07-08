// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"encoding/json"
	"fmt"
)

// contentType is the discriminator written into the JSON representation of a [Content] block so that a
// heterogeneous []Content can be deserialized again. encoding/json cannot unmarshal into an interface, so we
// tag every block with its concrete type and dispatch on it in [Message.UnmarshalJSON] and
// [ToolResult.UnmarshalJSON].
type contentType string

const (
	contentTypeText       contentType = "text"
	contentTypeMedia      contentType = "media"
	contentTypeFileRef    contentType = "file_ref"
	contentTypeToolCall   contentType = "tool_call"
	contentTypeToolResult contentType = "tool_result"
	contentTypeThinking   contentType = "thinking"
)

// contentEnvelope is the wire format of a single [Content] block: a discriminator plus the flattened payload
// of the concrete type. Using an envelope (instead of embedding the type into each struct) keeps the public
// content types free of persistence concerns.
type contentEnvelope struct {
	Type contentType `json:"type"`

	// text / thinking
	Text      string `json:"text,omitempty"`
	Signature string `json:"signature,omitempty"`

	// media
	Media *Media `json:"media,omitempty"`

	// file_ref
	FileRef *FileRef `json:"fileRef,omitempty"`

	// tool_call
	ToolCall *ToolCall `json:"toolCall,omitempty"`

	// tool_result
	ToolResult *toolResultWire `json:"toolResult,omitempty"`
}

// marshalContent converts a single [Content] into its envelope form.
func marshalContent(c Content) (contentEnvelope, error) {
	switch v := c.(type) {
	case Text:
		return contentEnvelope{Type: contentTypeText, Text: v.Text}, nil
	case Thinking:
		return contentEnvelope{Type: contentTypeThinking, Text: v.Text, Signature: v.Signature}, nil
	case Media:
		m := v
		return contentEnvelope{Type: contentTypeMedia, Media: &m}, nil
	case FileRef:
		fr := v
		return contentEnvelope{Type: contentTypeFileRef, FileRef: &fr}, nil
	case ToolCall:
		tc := v
		return contentEnvelope{Type: contentTypeToolCall, ToolCall: &tc}, nil
	case ToolResult:
		wire, err := toolResultToWire(v)
		if err != nil {
			return contentEnvelope{}, err
		}
		return contentEnvelope{Type: contentTypeToolResult, ToolResult: &wire}, nil
	default:
		return contentEnvelope{}, fmt.Errorf("completion: cannot marshal unknown content type %T", c)
	}
}

// unmarshalContent reconstructs the concrete [Content] from its envelope form.
func (e contentEnvelope) toContent() (Content, error) {
	switch e.Type {
	case contentTypeText:
		return Text{Text: e.Text}, nil
	case contentTypeThinking:
		return Thinking{Text: e.Text, Signature: e.Signature}, nil
	case contentTypeMedia:
		if e.Media == nil {
			return Media{}, nil
		}
		return *e.Media, nil
	case contentTypeFileRef:
		if e.FileRef == nil {
			return FileRef{}, nil
		}
		return *e.FileRef, nil
	case contentTypeToolCall:
		if e.ToolCall == nil {
			return ToolCall{}, nil
		}
		return *e.ToolCall, nil
	case contentTypeToolResult:
		if e.ToolResult == nil {
			return ToolResult{}, nil
		}
		return e.ToolResult.toToolResult()
	default:
		return nil, fmt.Errorf("completion: cannot unmarshal unknown content type %q", e.Type)
	}
}

// marshalContentSlice serializes a []Content into a slice of envelopes.
func marshalContentSlice(cs []Content) ([]contentEnvelope, error) {
	if cs == nil {
		return nil, nil
	}
	out := make([]contentEnvelope, 0, len(cs))
	for _, c := range cs {
		env, err := marshalContent(c)
		if err != nil {
			return nil, err
		}
		out = append(out, env)
	}
	return out, nil
}

// unmarshalContentSlice reconstructs a []Content from a slice of envelopes.
func unmarshalContentSlice(envs []contentEnvelope) ([]Content, error) {
	if envs == nil {
		return nil, nil
	}
	out := make([]Content, 0, len(envs))
	for _, env := range envs {
		c, err := env.toContent()
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// toolResultWire is the wire form of [ToolResult]. Its nested Content is itself a heterogeneous []Content and
// therefore carries envelopes too (tool results usually contain Text blocks, but the model may return media).
type toolResultWire struct {
	ToolCallID string            `json:"toolCallId"`
	Content    []contentEnvelope `json:"content,omitempty"`
	IsError    bool              `json:"isError,omitempty"`
}

func toolResultToWire(tr ToolResult) (toolResultWire, error) {
	inner, err := marshalContentSlice(tr.Content)
	if err != nil {
		return toolResultWire{}, err
	}
	return toolResultWire{
		ToolCallID: tr.ToolCallID,
		Content:    inner,
		IsError:    tr.IsError,
	}, nil
}

func (w toolResultWire) toToolResult() (ToolResult, error) {
	inner, err := unmarshalContentSlice(w.Content)
	if err != nil {
		return ToolResult{}, err
	}
	return ToolResult{
		ToolCallID: w.ToolCallID,
		Content:    inner,
		IsError:    w.IsError,
	}, nil
}

// messageWire is the wire form of [Message] with envelope-encoded content.
type messageWire struct {
	Role    Role              `json:"role"`
	Content []contentEnvelope `json:"content,omitempty"`
}

// MarshalJSON encodes the message with a type-tagged content representation so that [Message.UnmarshalJSON]
// can reconstruct the concrete [Content] blocks. This makes a []Message losslessly persistable (e.g. as part
// of a stored session).
func (m Message) MarshalJSON() ([]byte, error) {
	content, err := marshalContentSlice(m.Content)
	if err != nil {
		return nil, err
	}
	return json.Marshal(messageWire{Role: m.Role, Content: content})
}

// UnmarshalJSON is the inverse of [Message.MarshalJSON].
func (m *Message) UnmarshalJSON(data []byte) error {
	var wire messageWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	content, err := unmarshalContentSlice(wire.Content)
	if err != nil {
		return err
	}

	m.Role = wire.Role
	m.Content = content
	return nil
}

// MarshalJSON encodes a [ToolResult] using the envelope representation for its nested content. A ToolResult is
// also a [Content] block itself, so this keeps a standalone ToolResult consistent with the one nested inside a
// Message.
func (tr ToolResult) MarshalJSON() ([]byte, error) {
	wire, err := toolResultToWire(tr)
	if err != nil {
		return nil, err
	}
	return json.Marshal(wire)
}

// UnmarshalJSON is the inverse of [ToolResult.MarshalJSON].
func (tr *ToolResult) UnmarshalJSON(data []byte) error {
	var wire toolResultWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	res, err := wire.toToolResult()
	if err != nil {
		return err
	}

	*tr = res
	return nil
}
