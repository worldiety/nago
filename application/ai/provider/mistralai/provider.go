// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/provider"
)

var _ provider.Provider = (*mistralProvider)(nil)

type mistralProvider struct {
	id            provider.ID
	cfg           Settings
	cl            *Client
	libs          *mistralLibraries
	agents        *mistralAgents
	models        *mistralModels
	conversations *mistralConversations
}

func NewProvider(id provider.ID, cfg Settings) provider.Provider {
	p := &mistralProvider{
		id:  id,
		cfg: cfg,
		cl:  NewClient(cfg.Token),
	}

	p.libs = &mistralLibraries{
		parent: p,
	}

	p.agents = &mistralAgents{
		parent: p,
	}

	p.models = &mistralModels{
		parent: p,
	}

	p.conversations = &mistralConversations{
		parent: p,
	}

	return p
}

func (p *mistralProvider) Conversations() option.Opt[provider.Conversations] {
	return option.Some[provider.Conversations](p.conversations)
}

func (p *mistralProvider) Models() provider.Models {
	return p.models
}

func (p *mistralProvider) Identity() provider.ID {
	return p.id
}

func (p *mistralProvider) client() *Client {
	return p.cl
}

func (p *mistralProvider) Name() string {
	return p.cfg.Name
}

func (p *mistralProvider) Description() string {
	return p.cfg.Description
}

func (p *mistralProvider) Libraries() option.Opt[provider.Libraries] {
	return option.Some[provider.Libraries](p.libs)
}

func (p *mistralProvider) Agents() option.Opt[provider.Agents] {
	return option.Some[provider.Agents](p.agents)
}
