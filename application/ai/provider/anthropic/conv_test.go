// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"testing"

	"go.wdy.de/nago/application/ai/completion"
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
