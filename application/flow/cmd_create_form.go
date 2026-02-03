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

type CreateFormCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	Package     PackageID   `source:"nago.flow.packages"`
	Struct      TypeID      `source:"nago.flow.structs"`
	Name        Ident       `label:"nago.common.label.name"`
	Description string      `lines:"3"`
}

func (cmd CreateFormCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd CreateFormCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	if err := cmd.Name.Validate(); err != nil {
		errGrp.Add("Name", err.Error())
	}

	if !cmd.Name.IsPublic() {
		errGrp.Add("Name", "Name must start with an uppercase letter")
	}

	_, ok := ws.Packages.StructTypeByID(cmd.Struct)
	if !ok {
		errGrp.Add("Repository", "Repository not found")
		return nil, errGrp.Error()
	}

	if err := errGrp.Error(); err != nil {
		return nil, err
	}

	return []WorkspaceEvent{FormCreated{
		Workspace:   cmd.Workspace,
		ID:          data.RandIdent[FormID](),
		Struct:      cmd.Struct,
		Description: cmd.Description,
		Name:        cmd.Name,
		Package:     cmd.Package,
	}}, nil
}
