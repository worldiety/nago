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

type StringFieldAppended struct {
	Workspace   WorkspaceID `json:"workspace"`
	Struct      TypeID      `json:"struct"`
	Name        Ident       `json:"name"`
	ID          FieldID     `json:"id"`
	Description string      `json:"description"`
}

func (e StringFieldAppended) WorkspaceID() WorkspaceID {
	return e.Workspace
}

func (e StringFieldAppended) event() {}

type AppendStringFieldCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	Struct      TypeID      `source:"nago.flow.structs"`
	Name        Ident
	Description string `lines:"3"`
}

func (c AppendStringFieldCmd) WorkspaceID() WorkspaceID {
	return c.Workspace
}

func (c AppendStringFieldCmd) WithWorkspaceID(id WorkspaceID) AppendStringFieldCmd {
	c.Workspace = id
	return c
}

func NewAppendStringField(hnd handleCmd[StringFieldAppended]) AppendStringField {
	return func(subject auth.Subject, cmd AppendStringFieldCmd) (StringFieldAppended, error) {
		return hnd(subject, cmd.Workspace, func(ws *Workspace) (StringFieldAppended, error) {
			var zero StringFieldAppended
			s, ok := ws.structTypeByID(cmd.Struct)
			if !ok {
				return zero, xerrors.WithFields("Validation", "Struct", "Struct not found")
			}

			if err := cmd.Name.Validate(); err != nil {
				return zero, xerrors.WithFields("Validation", "Name", err.Error())
			}

			if !cmd.Name.IsPublic() {
				return zero, xerrors.WithFields("Validation", "Name", "Name must start with an uppercase letter")
			}

			for _, field := range s.fields {
				if field.Name() == cmd.Name {
					return zero, xerrors.WithFields("Validation", "Name", "Field already exists")
				}
			}

			id := data.RandIdent[FieldID]()

			return StringFieldAppended{
				Workspace:   cmd.Workspace,
				Struct:      cmd.Struct,
				ID:          id,
				Name:        cmd.Name,
				Description: cmd.Description,
			}, nil
		})
	}
}
