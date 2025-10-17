// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"log/slog"
	"sync"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xtime"
)

// Sync takes all mistral workspaces and tries to synchronize them into the mistral cloud. It does not make
// sense to apply a partial agent sync, because agents within a workspace may refer to each other (even in cycles) e.g.
// for handoffs.
type Sync func(subject auth.Subject) error

type UseCases struct {
	Sync Sync
}

type SynchronizedAgent struct {
	ID         agent.ID               `json:"id,omitempty"`
	CloudAgent string                 `json:"cloudAgent,omitempty"` // cloud agent id
	LastMod    xtime.UnixMilliseconds `json:"lastMod,omitempty"`    // last mod of the nago agent, which has been written into the cloud
}

func (a SynchronizedAgent) Identity() agent.ID {
	return a.ID
}

type SyncAgentRepository data.Repository[SynchronizedAgent, agent.ID]

func NewUseCases(bus events.Bus, repoWorkspaceName string, syncAgentRepo SyncAgentRepository, findSecret secret.Match, findWorkspaces workspace.FindWorkspacesByPlatform, findAgent agent.FindByID) UseCases {
	var mutex sync.Mutex
	uc := UseCases{
		Sync: NewSync(&mutex, repoWorkspaceName, syncAgentRepo, findSecret, findWorkspaces, findAgent),
	}

	events.SubscribeFor[agent.Updated](bus, func(evt agent.Updated) {
		if err := uc.Sync(user.SU()); err != nil {
			slog.Error("failed to sync mistral workspaces, caused by agent.Updated", "err", err.Error())
		}
	})

	return uc
}
