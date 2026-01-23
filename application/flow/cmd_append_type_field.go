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

type AppendTypeFieldCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	Struct      TypeID      `visible:"false" source:"nago.flow.structs"`
	Type        TypeID      `source:"nago.flow.types"`
	Name        Ident
	Description string `lines:"3"`
}

func (cmd AppendTypeFieldCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd AppendTypeFieldCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder
	s, ok := ws.Packages.StructTypeByID(cmd.Struct)
	if !ok {
		errGrp.Add("Struct", "Struct not found")
		return nil, errGrp.Error()
	}

	if err := cmd.Name.Validate(); err != nil {
		errGrp.Add("Name", err.Error())
	}

	if !cmd.Name.IsPublic() {
		errGrp.Add("Name", "Name must start with an uppercase letter")
	}

	if _, ok := s.Fields.ByName(cmd.Name); ok {
		errGrp.Add("Name", "Field already exists")
	}

	if _, ok := ws.Packages.TypeByID(cmd.Type); !ok {
		errGrp.Add("Type", "Type not found")
		return nil, errGrp.Error()
	}

	if err := errGrp.Error(); err != nil {
		return nil, err
	}

	id := data.RandIdent[FieldID]()

	return []WorkspaceEvent{TypeFieldAppended{
		Workspace:   cmd.Workspace,
		Struct:      cmd.Struct,
		ID:          id,
		Name:        cmd.Name,
		Description: cmd.Description,
		Type:        cmd.Type,
	}}, nil
}
