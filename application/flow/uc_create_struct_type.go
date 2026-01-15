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

type StructTypeCreated struct {
	Workspace   WorkspaceID `json:"workspace,omitempty"`
	ID          TypeID      `json:"id,omitempty"`
	Name        Ident       `json:"name,omitempty"`
	Package     PackageID   `json:"package,omitempty"`
	Description string      `json:"description,omitempty"`
}

func (e StructTypeCreated) WorkspaceID() WorkspaceID {
	return e.Workspace
}

func (e StructTypeCreated) event() {}

type CreateStructTypeCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	Package     PackageID   `source:"nago.flow.packages"`
	Name        Ident
	Description string
}

func (c CreateStructTypeCmd) WorkspaceID() WorkspaceID {
	return c.Workspace
}

func (c CreateStructTypeCmd) WithWorkspaceID(id WorkspaceID) CreateStructTypeCmd {
	c.Workspace = id
	return c
}

func NewCreateStructType(hnd handleCmd[StructTypeCreated]) CreateStructType {
	return func(subject auth.Subject, cmd CreateStructTypeCmd) (StructTypeCreated, error) {
		return hnd(subject, cmd.Workspace, func(ws *Workspace) (StructTypeCreated, error) {
			var zero StructTypeCreated
			_, ok := ws.packages[cmd.Package]
			if !ok {
				return zero, xerrors.WithFields("Validation", "Package not found")
			}

			if err := cmd.Name.Validate(); err != nil {
				return zero, xerrors.WithFields("Validation", "Name", err.Error())
			}

			if !cmd.Name.IsPublic() {
				return zero, xerrors.WithFields("Validation", "Name", "Name must start with an uppercase letter")
			}

			id := data.RandIdent[TypeID]()

			return StructTypeCreated{
				Workspace:   cmd.Workspace,
				ID:          id,
				Name:        cmd.Name,
				Package:     cmd.Package,
				Description: cmd.Description,
			}, nil
		})
	}
}
