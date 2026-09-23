// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package completion

import (
	"encoding/json"
	"errors"
	"testing"

	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/auth"
)

type lookIn struct {
	Selector string `json:"selector" optional:"true"`
}

func runOneCall(t *testing.T, tool Tool, args string) ToolResult {
	t.Helper()

	fake := &fakeCompletions{
		results: []Result{
			{
				Message:    Message{Role: Assistant, Content: []Content{ToolCall{ID: "1", Name: tool.Def.Name, Arguments: json.RawMessage(args)}}},
				StopReason: StopToolUse,
			},
			{
				Message:    Message{Role: Assistant, Content: []Content{Text{Text: "done"}}},
				StopReason: StopEndTurn,
			},
		},
	}

	_, history, err := Run(nil, fake, RunOptions{
		Options: Options{Messages: []Message{{Role: User, Content: []Content{Text{Text: "look"}}}}},
		Tools:   []Tool{tool},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if len(history) != 4 || len(history[2].Content) != 1 {
		t.Fatalf("unexpected history: %#v", history)
	}

	tr, ok := history[2].Content[0].(ToolResult)
	if !ok {
		t.Fatalf("expected a tool_result, got %T", history[2].Content[0])
	}

	return tr
}

// The content blocks, including inline image media, must end up inside the tool_result unchanged: that is the
// only place where a provider accepts an image as the answer to a tool call.
func TestRun_ContentTool_PlacesMediaInsideToolResult(t *testing.T) {
	var got lookIn
	tool := NewSubjectContentTool("look", "looks at the screen", func(_ auth.Subject, in lookIn) ([]Content, error) {
		got = in
		return []Content{
			Text{Text: "snapshot"},
			Media{MimeType: file.PNG, Source: Source{Data: []byte("png")}},
		}, nil
	})

	tr := runOneCall(t, tool, `{"selector":"#main"}`)

	if got.Selector != "#main" {
		t.Fatalf("arguments not decoded: %#v", got)
	}

	if tr.IsError || len(tr.Content) != 2 {
		t.Fatalf("unexpected tool_result: %#v", tr)
	}

	if txt, ok := tr.Content[0].(Text); !ok || txt.Text != "snapshot" {
		t.Fatalf("expected text first, got %#v", tr.Content[0])
	}

	media, ok := tr.Content[1].(Media)
	if !ok || media.MimeType != file.PNG || string(media.Source.Data) != "png" {
		t.Fatalf("expected inline png media, got %#v", tr.Content[1])
	}
}

func TestRun_ContentTool_ErrorIsReported(t *testing.T) {
	tool := NewSubjectContentTool("look", "looks at the screen", func(_ auth.Subject, in lookIn) ([]Content, error) {
		return nil, errors.New("frontend gone")
	})

	tr := runOneCall(t, tool, `{}`)
	if !tr.IsError {
		t.Fatalf("expected an error result, got %#v", tr)
	}
}

func TestRun_ContentTool_EmptyContentIsNotEmpty(t *testing.T) {
	tool := NewSubjectContentTool("look", "looks at the screen", func(_ auth.Subject, in lookIn) ([]Content, error) {
		return nil, nil
	})

	tr := runOneCall(t, tool, `{}`)
	if tr.IsError || len(tr.Content) != 1 {
		t.Fatalf("expected a single placeholder block, got %#v", tr)
	}
}
