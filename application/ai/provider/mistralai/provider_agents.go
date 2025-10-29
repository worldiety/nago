// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Agents = (*mistralAgents)(nil)

type mistralAgents struct {
	parent *mistralProvider
}

func (p *mistralAgents) FindByName(subject auth.Subject, name string) iter.Seq2[agent.Agent, error] {
	return func(yield func(agent.Agent, error) bool) {
		for a, err := range p.All(subject) {
			if err != nil {
				yield(a, err)
				return
			}

			if a.Name == name {
				if !yield(a, nil) {
					return
				}
			}
		}
		
	}
}

func (p *mistralAgents) FindByID(subject auth.Subject, id agent.ID) (option.Opt[agent.Agent], error) {
	a, err := p.client().GetAgent(string(id))
	if err != nil {
		return option.None[agent.Agent](), err
	}

	return option.Some(a.IntoAgent()), nil
}

func (p *mistralAgents) Create(subject auth.Subject, options agent.CreateOptions) (agent.Agent, error) {
	ag, err := p.client().CreateAgent(CreateAgentRequest{
		Model:        string(options.Model),
		Name:         options.Name,
		Description:  options.Description,
		Instructions: options.Instructions,
	})

	if err != nil {
		return agent.Agent{}, err
	}

	return ag.IntoAgent(), nil
}

func (p *mistralAgents) client() *Client {
	return p.parent.client()
}

func (p *mistralAgents) All(subject auth.Subject) iter.Seq2[agent.Agent, error] {
	return func(yield func(agent.Agent, error) bool) {
		for ag, err := range p.client().ListAgents() {
			if err != nil {
				yield(agent.Agent{}, err)
				return
			}

			if !yield(ag.IntoAgent(), nil) {
				return
			}
		}
	}
}

func (p *mistralAgents) Delete(subject auth.Subject, id agent.ID) error {
	return p.client().DeleteAgent(string(id))
}
