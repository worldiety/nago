// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/tool"
	"go.wdy.de/nago/auth"
)

var _ provider.Provider = (*anthropicProvider)(nil)

type anthropicProvider struct {
	id          provider.ID
	cfg         Settings
	cl          *Client
	models      *anthropicModels
	completions *anthropicCompletions
	files       *anthropicFiles
}

// NewProvider creates a stateless Anthropic (Claude) provider. Models, Tools, Completions and Files are
// supported; the remaining stateful capabilities (Libraries, Agents, Conversations) are intentionally
// unavailable because Anthropic's Messages API is stateless.
func NewProvider(id provider.ID, cfg Settings) provider.Provider {
	p := &anthropicProvider{
		id:  id,
		cfg: cfg,
		cl:  NewClient(cfg.Token, cfg.Version, cfg.RPS, cfg.Debug),
	}

	p.models = &anthropicModels{parent: p}
	p.completions = &anthropicCompletions{parent: p}
	p.files = &anthropicFiles{parent: p}

	return p
}

func (p *anthropicProvider) client() *Client {
	return p.cl
}

func (p *anthropicProvider) Identity() provider.ID {
	return p.id
}

func (p *anthropicProvider) Name() string {
	return p.cfg.Name
}

func (p *anthropicProvider) Description() string {
	return p.cfg.Description
}

func (p *anthropicProvider) Models() provider.Models {
	return p.models
}

func (p *anthropicProvider) Tools() provider.Tools {
	return anthropicTools{}
}

func (p *anthropicProvider) Completions() option.Opt[completion.Completions] {
	return option.Some[completion.Completions](p.completions)
}

// ----- intentionally unsupported stateful capabilities -----

func (p *anthropicProvider) Libraries() option.Opt[provider.Libraries] {
	return option.None[provider.Libraries]()
}

func (p *anthropicProvider) Agents() option.Opt[provider.Agents] {
	return option.None[provider.Agents]()
}

func (p *anthropicProvider) Conversations() option.Opt[provider.Conversations] {
	return option.None[provider.Conversations]()
}

func (p *anthropicProvider) Files() option.Opt[provider.Files] {
	return option.Some[provider.Files](p.files)
}

// anthropicTools reports no parameterless built-in tools. Function tools are supplied per request via
// completion.Options.Tools instead.
type anthropicTools struct{}

func (anthropicTools) All(subject auth.Subject) iter.Seq2[tool.Tool, error] {
	return func(yield func(tool.Tool, error) bool) {}
}

