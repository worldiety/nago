// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package conversation

import (
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xtime"
)

func NewStart(mutex *sync.Mutex, bus events.Bus, repo Repository, repoWS workspace.Repository, repoAgents agent.Repository) Start {
	return func(subject auth.Subject, opts StartOptions) (ID, error) {
		if !subject.HasPermission(PermStart) && !subject.HasResourcePermission(repoWS.Name(), string(opts.Workspace), PermStart) {
			return "", subject.Audit(PermStart)
		}

		mutex.Lock()
		defer mutex.Unlock()

		var err error
		var optWS option.Opt[workspace.Workspace]
		if opts.WorkspaceName != "" && opts.Workspace == "" {
			for ws, err := range repoWS.All() {
				if err != nil {
					return "", err
				}

				if ws.Name == opts.WorkspaceName {
					optWS = option.Some(ws)
					break
				}
			}

			if optWS.IsNone() {
				return "", fmt.Errorf("workspace by name %q not found", opts.WorkspaceName)
			}
		}

		if optWS.IsNone() {
			optWS, err = repoWS.FindByID(opts.Workspace)
			if err != nil || optWS.IsNone() {
				if err != nil {
					return "", fmt.Errorf("failed to find workspace %q: %w", opts.Workspace, err)
				}

				return "", fmt.Errorf("workspace is gone %q: %w", opts.Workspace, os.ErrNotExist)
			}
		}

		var optAgent option.Opt[agent.Agent]
		if opts.AgentName != "" && opts.Agent == "" {
			for ag, err := range repoAgents.All() {
				if err != nil {
					return "", err
				}

				if ag.Name == opts.AgentName {
					optAgent = option.Some(ag)
					break
				}
			}

			if optAgent.IsNone() {
				return "", fmt.Errorf("agent by name %q not found", opts.AgentName)
			}
		}

		if optAgent.IsNone() {
			if optAgent, err = repoAgents.FindByID(opts.Agent); err != nil || optAgent.IsSome() {
				if err != nil {
					return "", fmt.Errorf("failed to find agent %q: %w", opts.Workspace, err)
				}

				return "", fmt.Errorf("agent is gone %q: %w", opts.Workspace, os.ErrExist)
			}
		}

		if !slices.Contains(optWS.Unwrap().Agents, optAgent.Unwrap().ID) {
			return "", fmt.Errorf("agent is not in workspace %q.%q: %w", optWS.Unwrap().ID, optAgent.Unwrap().ID, os.ErrNotExist)
		}

		if len(opts.Input) == 0 {
			return "", fmt.Errorf("input must not be empty")
		}

		conv := Conversation{
			ID:           data.RandIdent[ID](),
			Workspace:    optWS.Unwrap().ID,
			Agent:        optAgent.Unwrap().ID,
			Name:         opts.Name,
			Description:  opts.Description,
			Instructions: opts.Instructions,
			Input:        opts.Input,
			CloudStore:   opts.CloudStore,
			CreatedAt:    xtime.Now(),
			CreatedBy:    subject.ID(),
		}

		if err := repo.Save(conv); err != nil {
			return "", fmt.Errorf("failed to save conversation: %w", err)
		}

		bus.Publish(Started{Conversation: conv.ID})

		return conv.ID, nil
	}
}
