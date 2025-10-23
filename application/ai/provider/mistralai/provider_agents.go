// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"iter"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ = provider.Agents((*mistralAgents)(nil))

type mistralAgents struct {
	parent *mistralProvider
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
