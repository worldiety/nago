// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
)

func baseOpts() completion.Options {
	return completion.Options{
		Model:  "claude-test",
		System: "you are a helpful assistant",
		Tools: []completion.ToolDef{
			{Name: "a", Schema: []byte(`{}`)},
			{Name: "b", Schema: []byte(`{}`)},
		},
		Messages: []completion.Message{
			{Role: completion.User, Content: []completion.Content{completion.Text{Text: "hi"}}},
			{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: "hello"}}},
			{Role: completion.User, Content: []completion.Content{completion.Text{Text: "new question"}}},
		},
	}
}

func TestBuildRequest_PromptCache_DefaultOn(t *testing.T) {
	p := &anthropicProvider{cfg: Settings{}}

	req, err := p.buildRequest(baseOpts())
	if err != nil {
		t.Fatal(err)
	}

	// last tool marked
	if cc := req.Tools[len(req.Tools)-1].CacheControl; cc == nil || cc.Type != "ephemeral" {
		t.Errorf("expected last tool to carry ephemeral cache_control, got %+v", cc)
	}
	// non-last tool not marked
	if req.Tools[0].CacheControl != nil {
		t.Errorf("expected first tool to be unmarked")
	}
	// system marked
	if cc := req.System[len(req.System)-1].CacheControl; cc == nil {
		t.Errorf("expected system block to carry cache_control")
	}
	// frozen history boundary == last block of second-to-last message
	prev := req.Messages[len(req.Messages)-2]
	if prev.Content[len(prev.Content)-1].CacheControl == nil {
		t.Errorf("expected second-to-last message to carry cache_control")
	}
	// newest message must NOT be cached
	last := req.Messages[len(req.Messages)-1]
	if last.Content[len(last.Content)-1].CacheControl != nil {
		t.Errorf("expected newest message to be uncached")
	}
}

func TestBuildRequest_PromptCache_Disabled(t *testing.T) {
	p := &anthropicProvider{cfg: Settings{DisablePromptCache: true}}

	req, err := p.buildRequest(baseOpts())
	if err != nil {
		t.Fatal(err)
	}

	for _, tl := range req.Tools {
		if tl.CacheControl != nil {
			t.Errorf("expected no cache_control on tools when disabled")
		}
	}
	for _, s := range req.System {
		if s.CacheControl != nil {
			t.Errorf("expected no cache_control on system when disabled")
		}
	}
	for _, m := range req.Messages {
		for _, c := range m.Content {
			if c.CacheControl != nil {
				t.Errorf("expected no cache_control on messages when disabled")
			}
		}
	}
}

func TestBuildRequest_PromptCache_TTL(t *testing.T) {
	p := &anthropicProvider{cfg: Settings{PromptCacheTTL: "1h"}}

	req, err := p.buildRequest(baseOpts())
	if err != nil {
		t.Fatal(err)
	}

	if cc := req.System[len(req.System)-1].CacheControl; cc == nil || cc.TTL != "1h" {
		t.Errorf("expected 1h ttl, got %+v", cc)
	}
}

// TestThinkingBlock_AlwaysSerializesThinkingField guards against a regression where an empty thinking
// text caused the "thinking" field to be dropped (json "omitempty"), which the stricter Anthropic schema
// rejects with `messages[i].content[j].thinking.thinking: Field required`.
func TestThinkingBlock_AlwaysSerializesThinkingField(t *testing.T) {
	cases := map[string]completion.Content{
		"non-empty": completion.Thinking{Text: "let me reason", Signature: "sig"},
		"empty":     completion.Thinking{Text: "", Signature: "sig"},
	}

	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			ac, err := toAPIContent(in)
			if err != nil {
				t.Fatal(err)
			}

			raw, err := json.Marshal(ac)
			if err != nil {
				t.Fatal(err)
			}

			var got map[string]any
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatal(err)
			}

			if got["type"] != "thinking" {
				t.Fatalf("expected type thinking, got %v", got["type"])
			}
			if _, ok := got["thinking"]; !ok {
				t.Errorf("thinking field must always be present, got %s", raw)
			}
			if _, ok := got["signature"]; !ok {
				t.Errorf("signature field must always be present, got %s", raw)
			}
		})
	}
}

// TestTextBlock_DoesNotLeakThinkingField ensures the shared content struct does not emit empty,
// type-irrelevant sibling fields (which the stricter schema may reject).
func TestTextBlock_DoesNotLeakThinkingField(t *testing.T) {
	ac, err := toAPIContent(completion.Text{Text: "hello"})
	if err != nil {
		t.Fatal(err)
	}

	raw, err := json.Marshal(ac)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(raw), "thinking") || strings.Contains(string(raw), "signature") {
		t.Errorf("text block must not contain thinking/signature fields, got %s", raw)
	}
}

func TestBuildRequest_PromptCache_SingleMessageNoHistoryBreakpoint(t *testing.T) {
	p := &anthropicProvider{cfg: Settings{}}

	opts := baseOpts()
	opts.Messages = opts.Messages[:1] // only one user message

	req, err := p.buildRequest(opts)
	if err != nil {
		t.Fatal(err)
	}

	if req.Messages[0].Content[0].CacheControl != nil {
		t.Errorf("expected no history breakpoint with a single message")
	}
}

// TestToolResult_InlineImageMediaStaysImage verifies that inline base64 image data (Media with Source.Data)
// is permitted as an image block inside a tool_result, since base64 sources are allowed there.
func TestToolResult_InlineImageMediaStaysImage(t *testing.T) {
	tr := completion.ToolResult{
		ToolCallID: "call-1",
		Content: []completion.Content{
			completion.Media{MimeType: file.PNG, Source: completion.Source{Data: []byte("pngbytes")}},
		},
	}

	ac, err := toAPIContent(tr)
	if err != nil {
		t.Fatal(err)
	}

	if len(ac.Content) != 1 || ac.Content[0].Type != "image" {
		t.Fatalf("expected single image block, got %+v", ac.Content)
	}
	if ac.Content[0].Source == nil || ac.Content[0].Source.Type != "base64" {
		t.Fatalf("expected base64 image source, got %+v", ac.Content[0].Source)
	}
}

// TestRequestUsesFileSource verifies the detector that decides whether the Files API beta header must be sent
// on a Messages request: it must be true iff some content block (including nested tool_result content) uses a
// file-id source, and false for inline base64/url-only requests.
func TestRequestUsesFileSource(t *testing.T) {
	// A file-id media source in a normal user message must trigger the beta.
	imgRef, err := toAPIContent(completion.Media{MimeType: file.PNG, Source: completion.Source{FileID: option.Some(file.ID("file-img"))}})
	if err != nil {
		t.Fatal(err)
	}
	withFile := apiRequest{Messages: []apiMessage{{Role: "user", Content: []apiContent{imgRef}}}}
	if !requestUsesFileSource(withFile) {
		t.Fatal("expected file source to be detected in a user message")
	}

	// Inline base64 image + plain text must NOT trigger the beta.
	inlineImg, err := toAPIContent(completion.Media{MimeType: file.PNG, Source: completion.Source{Data: []byte("x")}})
	if err != nil {
		t.Fatal(err)
	}
	withoutFile := apiRequest{Messages: []apiMessage{{Role: "user", Content: []apiContent{
		{Type: "text", Text: "hello"},
		inlineImg,
	}}}}
	if requestUsesFileSource(withoutFile) {
		t.Fatal("did not expect a file source for inline base64/text content")
	}

	// A file source nested inside a tool_result must also be detected.
	nested := apiRequest{Messages: []apiMessage{{Role: "user", Content: []apiContent{
		{Type: "tool_result", ToolUseID: "1", Content: []apiContent{
			{Type: "document", Source: &apiSource{Type: "file", FileID: "file-pdf"}},
		}},
	}}}}
	if !requestUsesFileSource(nested) {
		t.Fatal("expected file source nested in tool_result to be detected")
	}
}
