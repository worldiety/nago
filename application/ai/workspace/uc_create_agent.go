// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workspace

import (
	"fmt"
	"os"
	"sync"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xerrors"
)

func NewCreateAgent(mutex *sync.Mutex, repoWS Repository, repoAgents agent.Repository) CreateAgent {
	return func(subject auth.Subject, parent ID, createOptions CreateAgentOptions) (agent.ID, error) {
		if err := subject.AuditResource(repoWS.Name(), string(parent), PermCreateAgent); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optParent, err := repoWS.FindByID(parent)
		if err != nil {
			return "", err
		}

		if optParent.IsNone() {
			return "", fmt.Errorf("parent workspace is gone: %w", os.ErrNotExist)
		}

		ws := optParent.Unwrap()

		if createOptions.Name == "" {
			return "", xerrors.WithFields("invalid name", "Name", rstring.LabelValueMustNotBeEmpty.Get(subject))
		}

		if createOptions.Model == "" {
			createOptions.Model = agent.Balanced
		}

		aid := createOptions.System.ID
		if aid == "" {
			aid = data.RandIdent[agent.ID]()
		}

		if !createOptions.System.Valid {
			if optAgent, err := repoAgents.FindByID(aid); err != nil || optAgent.IsSome() {
				if err != nil {
					return "", err
				}

				return "", fmt.Errorf("agent id collision %q: %w", aid, os.ErrExist)
			}
		}

		ag := agent.Agent{
			ID:           aid,
			Name:         createOptions.Name,
			Description:  createOptions.Description,
			Prompt:       createOptions.Prompt,
			Model:        createOptions.Model,
			Libraries:    createOptions.Libraries,
			Capabilities: createOptions.Capabilities,
			Temperature:  createOptions.Temperature,
			System:       createOptions.System.Valid,
		}

		if err := repoAgents.Save(ag); err != nil {
			return "", fmt.Errorf("failed to save agent: %w", err)
		}

		ws.Agents = append(ws.Agents, ag.ID)
		if err := repoWS.Save(ws); err != nil {
			return "", fmt.Errorf("failed to save workspace: %w", err)
		}

		return ag.ID, nil
	}
}
