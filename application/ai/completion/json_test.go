// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/file"
)

// TestMessageJSONRoundtrip ensures a heterogeneous []Message survives a Marshal/Unmarshal cycle without loss,
// which is the precondition for persisting a completion history (e.g. inside a stored session).
func TestMessageJSONRoundtrip(t *testing.T) {
	history := []Message{
		{
			Role: User,
			Content: []Content{
				Text{Text: "please analyze this image and compute 3*12"},
				Media{
					MimeType: file.Type("image/png"),
					Source:   Source{FileID: option.Some(file.ID("file-123"))},
				},
			},
		},
		{
			Role: Assistant,
			Content: []Content{
				Thinking{Text: "I should use the calculator tool", Signature: "sig-abc"},
				ToolCall{ID: "call-1", Name: "calculator", Arguments: json.RawMessage(`{"op":"mul","a":3,"b":12}`)},
			},
		},
		{
			Role: User,
			Content: []Content{
				ToolResult{
					ToolCallID: "call-1",
					Content: []Content{
						Text{Text: "36"},
						Media{MimeType: file.Type("image/jpeg"), Source: Source{Data: []byte("binary")}},
						FileRef{File: file.ID("file-789"), MimeType: file.PDF},
					},
				},
				ToolResult{
					ToolCallID: "call-err",
					Content:    []Content{Text{Text: "boom"}},
					IsError:    true,
				},
			},
		},
		{
			Role:    Assistant,
			Content: []Content{Text{Text: "The result is 36."}},
		},
	}

	raw, err := json.Marshal(history)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got []Message
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !reflect.DeepEqual(history, got) {
		t.Fatalf("roundtrip mismatch:\nwant %#v\n got %#v\njson: %s", history, got, raw)
	}
}

// TestToolResultStandaloneRoundtrip verifies that a ToolResult marshalled on its own (not nested in a Message)
// is consistent with the nested representation, because ToolResult is both a Content block and carries nested
// content itself.
func TestToolResultStandaloneRoundtrip(t *testing.T) {
	tr := ToolResult{
		ToolCallID: "call-9",
		Content: []Content{
			Text{Text: "nested text"},
			ToolResult{ToolCallID: "inner", Content: []Content{Text{Text: "deep"}}},
		},
	}

	raw, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got ToolResult
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !reflect.DeepEqual(tr, got) {
		t.Fatalf("roundtrip mismatch:\nwant %#v\n got %#v\njson: %s", tr, got, raw)
	}
}

// TestUnknownContentTypeFails documents that an unknown discriminator is reported as an error instead of being
// silently dropped.
func TestUnknownContentTypeFails(t *testing.T) {
	var m Message
	err := json.Unmarshal([]byte(`{"role":"user","content":[{"type":"bogus"}]}`), &m)
	if err == nil {
		t.Fatal("expected error for unknown content type, got nil")
	}
}
