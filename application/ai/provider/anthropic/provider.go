// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/provider"
)

var _ provider.Provider = (*anthropicProvider)(nil)

type anthropicProvider struct {
	id          provider.ID
	cfg         Settings
	cl          *Client
	completions *anthropicCompletions
	files       *anthropicFiles

	// noAdaptiveThinking remembers models (model.ID as string) that rejected adaptive thinking, so the
	// automatic default is not sent to them again.
	noAdaptiveThinking sync.Map
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

func (p *anthropicProvider) Completions() option.Opt[completion.Completions] {
	return option.Some[completion.Completions](p.completions)
}

func (p *anthropicProvider) Files() option.Opt[provider.Files] {
	return option.Some[provider.Files](p.files)
}
