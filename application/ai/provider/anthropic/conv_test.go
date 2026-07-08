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

// TestFileRef_ImageBecomesImageBlockWithFileID verifies that an image FileRef is translated into an Anthropic
// image block sourced by file id — never inlined as base64.
func TestFileRef_ImageBecomesImageBlockWithFileID(t *testing.T) {
	ac, err := toAPIContent(completion.FileRef{File: "file-img", MimeType: file.PNG})
	if err != nil {
		t.Fatal(err)
	}

	if ac.Type != "image" {
		t.Fatalf("expected image block, got %q", ac.Type)
	}
	if ac.Source == nil || ac.Source.Type != "file" || ac.Source.FileID != "file-img" {
		t.Fatalf("expected file source with id, got %+v", ac.Source)
	}

	raw, _ := json.Marshal(ac)
	if strings.Contains(string(raw), "base64") || strings.Contains(string(raw), "\"data\"") {
		t.Errorf("file ref must not be inlined as base64, got %s", raw)
	}
}

// TestFileRef_DocumentOutsideToolResult verifies that a non-image FileRef becomes a document block when used
// outside a tool_result (where Anthropic permits documents).
func TestFileRef_DocumentOutsideToolResult(t *testing.T) {
	ac, err := toAPIContent(completion.FileRef{File: "file-pdf", MimeType: file.PDF})
	if err != nil {
		t.Fatal(err)
	}

	if ac.Type != "document" {
		t.Fatalf("expected document block, got %q", ac.Type)
	}
	if ac.Source == nil || ac.Source.FileID != "file-pdf" {
		t.Fatalf("expected file source with id, got %+v", ac.Source)
	}
}

// TestToolResult_NonImageFileRefIsTextNotDocument guards the core invariant: Anthropic rejects document
// blocks inside a tool_result, so a non-image FileRef there must be degraded to a text reference (keeping the
// file id) instead of a document block, and must never be inlined as base64.
func TestToolResult_NonImageFileRefIsTextNotDocument(t *testing.T) {
	tr := completion.ToolResult{
		ToolCallID: "call-1",
		Content: []completion.Content{
			completion.Text{Text: "done"},
			completion.FileRef{File: "file-pdf", MimeType: file.PDF},
		},
	}

	ac, err := toAPIContent(tr)
	if err != nil {
		t.Fatal(err)
	}

	if ac.Type != "tool_result" {
		t.Fatalf("expected tool_result, got %q", ac.Type)
	}
	if len(ac.Content) != 2 {
		t.Fatalf("expected 2 nested blocks, got %d", len(ac.Content))
	}
	for _, nested := range ac.Content {
		if nested.Type == "document" {
			t.Fatalf("tool_result must not contain a document block: %+v", ac.Content)
		}
	}
	if ac.Content[1].Type != "text" {
		t.Fatalf("expected non-image file ref to become text, got %q", ac.Content[1].Type)
	}
	if !strings.Contains(ac.Content[1].Text, "file-pdf") {
		t.Errorf("expected text reference to carry the file id, got %q", ac.Content[1].Text)
	}
}

// TestToolResult_ImageFileRefBecomesText verifies that an image FileRef inside a tool_result is degraded to a
// text reference: Anthropic rejects file-id sources on image blocks inside a tool_result (only base64/url are
// allowed there), so a FileRef — which is always a file id — must never be emitted as an image block here. The
// text keeps the file id so the model can attach the image in a later user turn.
func TestToolResult_ImageFileRefBecomesText(t *testing.T) {
	tr := completion.ToolResult{
		ToolCallID: "call-1",
		Content: []completion.Content{
			completion.FileRef{File: "file-img", MimeType: file.JPEG},
		},
	}

	ac, err := toAPIContent(tr)
	if err != nil {
		t.Fatal(err)
	}

	if len(ac.Content) != 1 {
		t.Fatalf("expected 1 nested block, got %d", len(ac.Content))
	}
	if ac.Content[0].Type != "text" {
		t.Fatalf("expected image file ref to become text inside tool_result, got %q", ac.Content[0].Type)
	}
	if ac.Content[0].Source != nil {
		t.Fatalf("text block must not carry a source, got %+v", ac.Content[0].Source)
	}
	if !strings.Contains(ac.Content[0].Text, "file-img") {
		t.Errorf("expected text reference to carry the file id, got %q", ac.Content[0].Text)
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
	// A FileRef in a normal user message resolves to a file source and must trigger the beta.
	imgRef, err := toAPIContent(completion.FileRef{File: "file-img", MimeType: file.PNG})
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
