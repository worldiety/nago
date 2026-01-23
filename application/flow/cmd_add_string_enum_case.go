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

type AddStringEnumCaseCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	String      TypeID      `visible:"false"`
	Name        Ident
	Value       string
	Description string `lines:"3"`
}

func (cmd AddStringEnumCaseCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd AddStringEnumCaseCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder
	s, ok := ws.Packages.StringTypeByID(cmd.String)
	if !ok {
		errGrp.Add("String", "Type not found")
		return nil, errGrp.Error()
	}

	if err := cmd.Name.Validate(); err != nil {
		errGrp.Add("Name", err.Error())
	}

	if !cmd.Name.IsPublic() {
		errGrp.Add("Name", "Name must start with an uppercase letter")
	}

	if _, ok := s.Enumeration.ByName(cmd.Name); ok {
		errGrp.Add("Name", "Case already exists")
	}

	if _, ok := s.Enumeration.ByValue(cmd.Value); ok {
		errGrp.Add("Value", "Value already exists")
	}

	return []WorkspaceEvent{StringEnumCaseAdded{
		Workspace:   cmd.Workspace,
		String:      cmd.String,
		Name:        cmd.Name,
		Value:       cmd.Value,
		Description: cmd.Description,
	}}, nil
}
