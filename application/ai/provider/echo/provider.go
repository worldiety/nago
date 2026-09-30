// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package echo provides a provider without any capability, e.g. for tests which only need a provider identity.
package echo

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/provider"
)

type Provider struct {
	id   provider.ID
	name string
}

var _ provider.Provider = (*Provider)(nil)

func New(id provider.ID, name string) *Provider {
	return &Provider{id: id, name: name}
}

func (p *Provider) Identity() provider.ID {
	return p.id
}

func (p *Provider) Name() string {
	return p.name
}

func (p *Provider) Description() string {
	return ""
}

func (p *Provider) Files() option.Opt[provider.Files] {
	return option.Opt[provider.Files]{}
}

func (p *Provider) Completions() option.Opt[completion.Completions] {
	return option.Opt[completion.Completions]{}
}
