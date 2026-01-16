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

type BoolFieldAppended struct {
	Workspace   WorkspaceID `json:"workspace"`
	Struct      TypeID      `json:"struct"`
	Name        Ident       `json:"name"`
	ID          FieldID     `json:"id"`
	Description string      `json:"description"`
}

func (e BoolFieldAppended) WorkspaceID() WorkspaceID {
	return e.Workspace
}

func (e BoolFieldAppended) event() {}

type AppendBoolFieldCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	Struct      TypeID      `visible:"false" source:"nago.flow.structs"`
	Name        Ident
	Description string `lines:"3"`
}

func (c AppendBoolFieldCmd) cmd() {}

func NewAppendBoolField(hnd handleCmd[BoolFieldAppended]) AppendBoolField {
	return func(subject auth.Subject, cmd AppendBoolFieldCmd) (BoolFieldAppended, error) {
		return hnd(subject, cmd.Workspace, func(ws *Workspace) (BoolFieldAppended, error) {
			var zero BoolFieldAppended
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

			return BoolFieldAppended{
				Workspace:   cmd.Workspace,
				Struct:      cmd.Struct,
				ID:          id,
				Name:        cmd.Name,
				Description: cmd.Description,
			}, nil
		})
	}
}
