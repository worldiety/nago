// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

type FindWorkspaces func(subject auth.Subject) iter.Seq2[WorkspaceID, error]
type LoadWorkspace func(subject auth.Subject, id WorkspaceID) (option.Opt[*Workspace], error)

type CreateWorkspace func(subject auth.Subject, cmd CreateWorkspaceCmd) (WorkspaceID, error)

type UseCases struct {
	FindWorkspaces  FindWorkspaces
	LoadWorkspace   LoadWorkspace
	CreateWorkspace CreateWorkspace
}

func NewUseCases(repoName string, storeEvent evs.Store[WorkspaceEvent], replayWorkspace evs.ReplayWithIndex[WorkspaceID, WorkspaceEvent], wsIndex *evs.StoreIndex[WorkspaceID, WorkspaceEvent]) UseCases {
	return UseCases{
		FindWorkspaces:  NewFindWorkspace(repoName, wsIndex),
		LoadWorkspace:   NewLoadWorkspace(repoName, replayWorkspace),
		CreateWorkspace: NewCreateWorkspace(storeEvent),
	}
}
