// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xerrors"
)

type PrimaryKeySelected struct {
	Workspace WorkspaceID `json:"workspace"`
	Struct    TypeID      `json:"struct"`
	Field     FieldID     `json:"field"`
}

func (e PrimaryKeySelected) WorkspaceID() WorkspaceID {
	return e.Workspace
}

func (e PrimaryKeySelected) event() {}

type SelectPrimaryKeyCmd struct {
	Workspace WorkspaceID `visible:"false"`
	Struct    TypeID      `source:"nago.flow.pkstructs"`
	Field     FieldID     `source:"nago.flow.self.pkcandidates"`
}

func (c SelectPrimaryKeyCmd) cmd() {}

type SelectPrimaryKey func(subject auth.Subject, cmd SelectPrimaryKeyCmd) (PrimaryKeySelected, error)

func NewSelectPrimaryKey(hnd handleCmd[PrimaryKeySelected]) SelectPrimaryKey {
	return func(subject auth.Subject, cmd SelectPrimaryKeyCmd) (PrimaryKeySelected, error) {
		return hnd(subject, cmd.Workspace, func(ws *Workspace) (PrimaryKeySelected, error) {
			var zero PrimaryKeySelected

			s, ok := ws.structTypeByID(cmd.Struct)
			if !ok {
				return zero, xerrors.WithFields("Validation", "Struct", "Struct not found")
			}

			var pkField Field
			for _, field := range s.fields {
				if field.Identity() == cmd.Field {
					pkField = field
				}
			}

			if pkField == nil {
				return zero, xerrors.WithFields("Validation", "Field", "Field not found")
			}

			if _, ok := pkField.(*StringField); !ok {
				return zero, xerrors.WithFields("Validation", "Field", "Field is not a string field")
			}

			return PrimaryKeySelected{
				Workspace: cmd.Workspace,
				Struct:    cmd.Struct,
				Field:     pkField.Identity(),
			}, nil
		})
	}
}
