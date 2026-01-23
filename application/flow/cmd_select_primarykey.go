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

type SelectPrimaryKeyCmd struct {
	Workspace WorkspaceID `visible:"false"`
	Struct    TypeID      `source:"nago.flow.pkstructs"`
	Field     FieldID     `source:"nago.flow.self.pkcandidates"`
}

func (cmd SelectPrimaryKeyCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd SelectPrimaryKeyCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {

	var errGrp xerrors.FieldBuilder
	s, ok := ws.Packages.StructTypeByID(cmd.Struct)
	if !ok {
		errGrp.Add("Struct", "Struct not found")
		return nil, errGrp.Error()
	}

	targetField, ok := s.Fields.ByID(cmd.Field)
	if !ok {
		errGrp.Add("Field", "Field not found")
		return nil, errGrp.Error()
	}

	pkField, ok := targetField.(PKField)
	if ok && !pkField.SuitableAsPrimaryKey(ws) {
		errGrp.Add("Field", "Field is not suitable as primary key")
		return nil, errGrp.Error()
	}

	return []WorkspaceEvent{PrimaryKeySelected{
		Workspace: cmd.Workspace,
		Struct:    cmd.Struct,
		Field:     targetField.Identity(),
	}}, nil
}
