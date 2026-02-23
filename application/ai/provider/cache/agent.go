// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cache

import (
	"fmt"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

type cacheAgent struct {
	parent *Provider
	id     agent.ID
}

func (c cacheAgent) Identity() agent.ID {
	return c.id
}

func (c cacheAgent) Update(subject auth.Subject, opts agent.UpdateOptions) (agent.Agent, error) {
	optAg, err := c.parent.repoAgents.FindByID(c.id)
	if err != nil {
		return agent.Agent{}, err
	}

	if optAg.IsNone() {
		return agent.Agent{}, fmt.Errorf("no such agent: %s", c.id)
	}

	ag := optAg.Unwrap()

	if ag.CreatedBy != subject.ID() && !subject.HasResourcePermission(rebac.Namespace(c.parent.repoAgents.Name()), rebac.Instance(c.id), PermAgentUpdate) {
		return agent.Agent{}, subject.Audit(PermAgentUpdate)
	}

	ag, err = c.parent.prov.Agents().Unwrap().Agent(c.id).Update(subject, opts)
	if err != nil {
		return agent.Agent{}, err
	}

	ag.CreatedBy = subject.ID()
	if ag.UpdatedAt == 0 {
		ag.UpdatedAt = xtime.Now()
	}

	if err := c.parent.repoAgents.Save(ag); err != nil {
		return agent.Agent{}, err
	}

	return ag, nil
}
