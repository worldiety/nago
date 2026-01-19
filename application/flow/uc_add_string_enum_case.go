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

type StringEnumCaseAdded struct {
	Workspace   WorkspaceID `json:"workspace"`
	String      TypeID      `json:"string"`
	Name        Ident       `json:"name"`
	Value       string      `json:"value"`
	Description string      `json:"description"`
}

func (e StringEnumCaseAdded) WorkspaceID() WorkspaceID {
	return e.Workspace
}

func (e StringEnumCaseAdded) event() {}

type AddStringEnumCaseCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	String      TypeID      `visible:"false"`
	Name        Ident
	Value       string
	Description string `lines:"3"`
}

func (c AddStringEnumCaseCmd) cmd() {}

type AddStringEnumCase func(subject auth.Subject, cmd AddStringEnumCaseCmd) (StringEnumCaseAdded, error)

func NewAddStringEnumCase(hnd handleCmd[StringEnumCaseAdded]) AddStringEnumCase {
	return func(subject auth.Subject, cmd AddStringEnumCaseCmd) (StringEnumCaseAdded, error) {
		return hnd(subject, cmd.Workspace, func(ws *Workspace) (StringEnumCaseAdded, error) {
			var zero StringEnumCaseAdded

			s, ok := ws.stringTypeByID(cmd.String)
			if !ok {
				return zero, xerrors.WithFields("Validation", "String", "Type not found")
			}

			if err := cmd.Name.Validate(); err != nil {
				return zero, xerrors.WithFields("Validation", "Name", err.Error())
			}

			if !cmd.Name.IsPublic() {
				return zero, xerrors.WithFields("Validation", "Name", "Name must start with an uppercase letter")
			}

			for eCase := range s.values.Load().All() {
				if eCase.name == cmd.Name {
					return zero, xerrors.WithFields("Validation", "Name", "Case already exists")
				}

				if (*eCase.value.Load()) == cmd.Value {
					return zero, xerrors.WithFields("Validation", "Value", "Value already exists")
				}
			}

			return StringEnumCaseAdded{
				Workspace:   cmd.Workspace,
				String:      cmd.String,
				Name:        cmd.Name,
				Value:       cmd.Value,
				Description: cmd.Description,
			}, nil
		})
	}
}
