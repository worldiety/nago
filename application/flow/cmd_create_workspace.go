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
	Name        Ident  `label:"nago.common.label.name"`
	Description string `label:"nago.common.label.description" lines:"3"`
}

func (cmd CreateWorkspaceCmd) WorkspaceID() WorkspaceID {
	return ""
}

func (cmd CreateWorkspaceCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder
	if err := cmd.Name.Validate(); err != nil {
		errGrp.Add("Name", err.Error())
	}

	if cmd.Name.IsPublic() {
		errGrp.Add("Name", "Name must not start with an uppercase letter")
	}

	if err := errGrp.Error(); err != nil {
		return nil, err
	}

	id := data.RandIdent[WorkspaceID]()
	evt := WorkspaceCreated{
		Workspace:   id,
		Name:        cmd.Name,
		Description: cmd.Description,
	}

	return []WorkspaceEvent{evt}, nil
}
