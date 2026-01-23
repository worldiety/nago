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

type CreateStringTypeCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	Package     PackageID   `source:"nago.flow.packages"`
	Name        Ident
	Description string
}

func (cmd CreateStringTypeCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd CreateStringTypeCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder
	if _, ok := ws.Packages.ByID(cmd.Package); !ok {
		errGrp.Add("Package", "Package not found")
	}

	if err := cmd.Name.Validate(); err != nil {
		errGrp.Add("Name", err.Error())
	}

	if !cmd.Name.IsPublic() {
		errGrp.Add("Name", "Name must start with an uppercase letter")
	}

	if err := errGrp.Error(); err != nil {
		return nil, err
	}

	id := data.RandIdent[TypeID]()

	return []WorkspaceEvent{StringTypeCreated{
		Workspace:   cmd.Workspace,
		ID:          id,
		Name:        cmd.Name,
		Package:     cmd.Package,
		Description: cmd.Description,
	}}, nil
}
