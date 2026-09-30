// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package openai

import (
	"sync/atomic"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/provider"
)

var _ provider.Provider = (*openaiProvider)(nil)

type openaiProvider struct {
	id          provider.ID
	cfg         Settings
	cl          *Client
	completions *openaiCompletions

	// useMaxCompletionTokens selects max_completion_tokens instead of the legacy max_tokens field. It starts
	// true for api.openai.com (whose reasoning models reject max_tokens) and false for compatible servers (which
	// often know only max_tokens), and flips once a server rejects the chosen field.
	useMaxCompletionTokens atomic.Bool

	// noStreamUsage remembers that the server rejected stream_options, so usage is not requested any more.
	noStreamUsage atomic.Bool
}

// NewProvider creates a stateless provider for the OpenAI Chat Completions API. It works with OpenAI and
// with OpenAI-compatible servers (Ollama, vLLM, LM Studio, OpenRouter, llama.cpp server, ...), selected by
// [Settings.BaseURL]. Only completions are supported; Files is unavailable, because compatible servers lack a
// files API, so binary content is always sent inline.
func NewProvider(id provider.ID, cfg Settings) provider.Provider {
	return newProvider(id, cfg)
}

func newProvider(id provider.ID, cfg Settings) *openaiProvider {
	p := &openaiProvider{
		id:  id,
		cfg: cfg,
		cl:  NewClient(cfg.BaseURL, cfg.Token, cfg.RPS, cfg.Debug),
	}

	p.useMaxCompletionTokens.Store(p.cl.isOpenAI())
	p.completions = &openaiCompletions{parent: p}

	return p
}

func (p *openaiProvider) client() *Client {
	return p.cl
}

func (p *openaiProvider) Identity() provider.ID {
	return p.id
}

func (p *openaiProvider) Name() string {
	return p.cfg.Name
}

func (p *openaiProvider) Description() string {
	return p.cfg.Description
}

func (p *openaiProvider) Completions() option.Opt[completion.Completions] {
	return option.Some[completion.Completions](p.completions)
}

func (p *openaiProvider) Files() option.Opt[provider.Files] {
	return option.None[provider.Files]()
}
