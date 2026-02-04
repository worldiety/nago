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

type DeleteWorkspace func(subject auth.Subject, id WorkspaceID) error
type HandleCommand func(subject auth.Subject, cmd WorkspaceCommand) error

// ExportedWorkspace represents a json decoded Export. See also [ExportWorkspace] and [ImportWorkspace].
type ExportedWorkspace struct {
	Events []evs.JsonEnvelope `json:"events"`
}
type ExportWorkspace func(subject auth.Subject, id WorkspaceID) ([]byte, error)
type ImportWorkspace func(subject auth.Subject, data []byte) error

type UseCases struct {
	FindWorkspaces  FindWorkspaces
	LoadWorkspace   LoadWorkspace
	HandleCommand   HandleCommand
	DeleteWorkspace DeleteWorkspace
	ExportWorkspace ExportWorkspace
	ImportWorkspace ImportWorkspace
}

func NewUseCases(repoName string, handler *evs.Handler[*Workspace, WorkspaceEvent, WorkspaceID], wsIndex *evs.StoreIndex[WorkspaceID, WorkspaceEvent]) UseCases {

	return UseCases{
		FindWorkspaces:  NewFindWorkspace(repoName, wsIndex),
		LoadWorkspace:   NewLoadWorkspace(repoName, handler),
		HandleCommand:   NewHandleCommand(handler),
		DeleteWorkspace: NewDeleteWorkspace(handler),
		ExportWorkspace: NewExportWorkspace(handler),
		ImportWorkspace: NewImportWorkspace(handler),
	}
}
