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

type HandleCommand func(subject auth.Subject, cmd WorkspaceCommand) error
type UseCases struct {
	FindWorkspaces FindWorkspaces
	LoadWorkspace  LoadWorkspace
	HandleCommand  HandleCommand
}

func NewUseCases(repoName string, handler *evs.Handler[*Workspace, WorkspaceEvent, WorkspaceID], wsIndex *evs.StoreIndex[WorkspaceID, WorkspaceEvent]) UseCases {

	return UseCases{
		FindWorkspaces: NewFindWorkspace(repoName, wsIndex),
		LoadWorkspace:  NewLoadWorkspace(repoName, handler),
		HandleCommand:  NewHandleCommand(handler),
	}
}
