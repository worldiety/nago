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

type TypeFieldAppended struct {
	Workspace   WorkspaceID `json:"workspace"`
	Struct      TypeID      `json:"struct"`
	Name        Ident       `json:"name"`
	ID          FieldID     `json:"id"`
	Type        TypeID      `json:"type"`
	Description string      `json:"description"`
}

func (e TypeFieldAppended) WorkspaceID() WorkspaceID {
	return e.Workspace
}

func (e TypeFieldAppended) event() {}

type AppendTypeFieldCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	Struct      TypeID      `visible:"false" source:"nago.flow.structs"`
	Type        TypeID      `source:"nago.flow.types"`
	Name        Ident
	Description string `lines:"3"`
}

func (c AppendTypeFieldCmd) cmd() {}

type AppendTypeField func(subject auth.Subject, cmd AppendTypeFieldCmd) (TypeFieldAppended, error)

func NewAppendTypeField(hnd handleCmd[TypeFieldAppended]) AppendTypeField {
	return func(subject auth.Subject, cmd AppendTypeFieldCmd) (TypeFieldAppended, error) {
		return hnd(subject, cmd.Workspace, func(ws *Workspace) (TypeFieldAppended, error) {
			var zero TypeFieldAppended
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

			if _, ok := ws.typeByID(cmd.Type); !ok {
				return zero, xerrors.WithFields("Validation", "Type", "Type not found")
			}

			id := data.RandIdent[FieldID]()

			return TypeFieldAppended{
				Workspace:   cmd.Workspace,
				Struct:      cmd.Struct,
				ID:          id,
				Name:        cmd.Name,
				Description: cmd.Description,
				Type:        cmd.Type,
			}, nil
		})
	}
}
