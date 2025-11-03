// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cache

import (
	"context"
	"fmt"
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

var _ provider.Agents = (*cacheAgents)(nil)

type cacheAgents struct {
	parent *Provider
}

func (c cacheAgents) Agent(id agent.ID) provider.Agent {
	return cacheAgent{c.parent, id}
}

func (c cacheAgents) All(subject auth.Subject) iter.Seq2[agent.Agent, error] {
	return func(yield func(agent.Agent, error) bool) {
		for key, err := range c.parent.idxProvAgents.AllByPrimary(context.Background(), c.parent.Identity()) {
			if err != nil {
				if !yield(agent.Agent{}, err) {
					return
				}

				continue
			}

			optConv, err := c.parent.repoAgents.FindByID(key.Secondary)
			if err != nil {
				if !yield(agent.Agent{}, err) {
					return
				}

				continue
			}

			if optConv.IsNone() {
				continue // stale ref
			}

			m := optConv.Unwrap()

			if m.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoModels.Name(), string(m.ID), PermAgentFindAll) {
				continue
			}

			if !yield(m, nil) {
				return
			}
		}
	}
}

func (c cacheAgents) Delete(subject auth.Subject, id agent.ID) error {
	optAg, err := c.parent.repoAgents.FindByID(id)
	if err != nil {
		return err
	}

	if optAg.IsNone() {
		return nil
	}

	ag := optAg.Unwrap()
	if ag.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoAgents.Name(), string(ag.ID), PermAgentDelete) {
		return subject.Audit(PermDocumentDelete)
	}

	if err := c.parent.prov.Agents().Unwrap().Delete(subject, id); err != nil {
		return err
	}

	if err := c.parent.idxProvAgents.Delete(context.Background(), c.parent.Identity(), ag.ID); err != nil {
		return err
	}

	return c.parent.repoAgents.DeleteByID(id)
}

func (c cacheAgents) FindByID(subject auth.Subject, id agent.ID) (option.Opt[agent.Agent], error) {
	optAg, err := c.parent.repoAgents.FindByID(id)
	if err != nil {
		return option.Opt[agent.Agent]{}, err
	}

	if optAg.IsNone() {
		return option.Opt[agent.Agent]{}, nil
	}

	ag := optAg.Unwrap()
	if ag.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoAgents.Name(), string(ag.ID), PermAgentFindByID) {
		return option.Opt[agent.Agent]{}, subject.Audit(PermAgentFindByID)
	}

	return optAg, nil
}

func (c cacheAgents) FindByName(subject auth.Subject, name string) iter.Seq2[agent.Agent, error] {
	return func(yield func(agent.Agent, error) bool) {
		for ag, err := range c.parent.repoAgents.All() {
			if err != nil {
				if !yield(ag, err) {
					return
				}

				continue
			}

			if ag.Name == ag.Name && (ag.CreatedBy == subject.ID() || subject.HasPermission(PermAgentFindByName)) {
				if !yield(ag, nil) {
					return
				}
			}
		}
	}
}

func (c cacheAgents) Create(subject auth.Subject, opts agent.CreateOptions) (agent.Agent, error) {
	if err := subject.Audit(PermAgentCreate); err != nil {
		return agent.Agent{}, err
	}

	ag, err := c.parent.prov.Agents().Unwrap().Create(subject, opts)
	if err != nil {
		return agent.Agent{}, err
	}

	if ag.UpdatedAt == 0 {
		ag.UpdatedAt = xtime.Now()
	}

	ag.CreatedBy = subject.ID()
	if ag.Identity() == "" {
		return agent.Agent{}, fmt.Errorf("provider returned empty identity")
	}

	if opt, err := c.parent.repoAgents.FindByID(ag.ID); err != nil || opt.IsSome() {
		if err != nil {
			return agent.Agent{}, err
		}

		return agent.Agent{}, fmt.Errorf("provider returned an existing agent: %s", ag.ID)
	}

	if err := c.parent.repoAgents.Save(ag); err != nil {
		return agent.Agent{}, err
	}

	if err := c.parent.idxProvAgents.Put(c.parent.Identity(), ag.ID); err != nil {
		return agent.Agent{}, err
	}

	return ag, nil
}
