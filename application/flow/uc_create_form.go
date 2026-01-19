// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xerrors"
)

type FormID string

type FormCreated struct {
	Workspace   WorkspaceID  `json:"workspace"`
	Repository  RepositoryID `json:"repository"`
	ID          FormID       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
}

func (e FormCreated) WorkspaceID() WorkspaceID {
	return e.Workspace
}

func (e FormCreated) event() {}

type CreateFormCmd struct {
	Workspace   WorkspaceID  `visible:"false"`
	Repository  RepositoryID `source:"nago.flow.repositories"`
	Name        string       `label:"nago.common.label.name"`
	Description string       `lines:"3"`
}

func (c CreateFormCmd) cmd() {}

type CreateForm func(subject auth.Subject, cmd CreateFormCmd) (FormCreated, error)

func NewCreateForm(hnd handleCmd[FormCreated]) CreateForm {
	return func(subject auth.Subject, cmd CreateFormCmd) (FormCreated, error) {
		return hnd(subject, cmd.Workspace, func(ws *Workspace) (FormCreated, error) {
			var zero FormCreated
			if err := cmd.Repository.Validate(); err != nil {
				return zero, xerrors.WithFields("Validation", "Repository", err.Error())
			}

			_, ok := ws.repositoryByID(cmd.Repository)
			if !ok {
				return zero, xerrors.WithFields("Validation", "Repository", "Repository not found")
			}

			return FormCreated{
				Workspace:   cmd.Workspace,
				ID:          data.RandIdent[FormID](),
				Repository:  cmd.Repository,
				Description: cmd.Description,
				Name:        cmd.Name,
			}, nil
		})
	}
}
