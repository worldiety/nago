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

type CreateWorkspaceCmd struct {
	Label       string
	Name        Ident       `label:"nago.common.label.name"`
	Description string      `label:"nago.common.label.description" lines:"3"`
	ID          WorkspaceID `visible:"false"`
}

func (cmd CreateWorkspaceCmd) WorkspaceID() WorkspaceID {
	return cmd.ID
}

func (cmd CreateWorkspaceCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder
	if err := cmd.Name.Validate(); err != nil {
		errGrp.Add("Name", err.Error())
	}

	if cmd.Name.IsPublic() {
		errGrp.Add("Name", "Name must not start with an uppercase letter")
	}

	if cmd.Label == "" {
		errGrp.Add("Label", "Label must not be empty")
	}

	if err := errGrp.Error(); err != nil {
		return nil, err
	}

	id := cmd.ID
	if id == "" {
		id = data.RandIdent[WorkspaceID]()
	}

	evt := WorkspaceCreated{
		Workspace:   id,
		Name:        cmd.Name,
		Description: cmd.Description,
	}

	cpkg := PackageCreated{
		Workspace:   evt.Workspace,
		Package:     data.RandIdent[PackageID](),
		Path:        ImportPath(evt.Name),
		Name:        evt.Name,
		Description: "Default workspace package.",
	}

	return []WorkspaceEvent{evt, cpkg}, nil
}
