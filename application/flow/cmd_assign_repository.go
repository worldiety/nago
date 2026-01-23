// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"slices"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xerrors"
)

type AssignRepositoryCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	Struct      TypeID      `source:"nago.flow.pkstructs"`
	Repository  RepositoryID
	Description string `lines:"3"`
}

func (cmd AssignRepositoryCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd AssignRepositoryCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	if err := cmd.Repository.Validate(); err != nil {
		errGrp.Add("Repository", err.Error())
	}

	if _, ok := ws.Repositories.ByID(cmd.Repository); ok {
		errGrp.Add("Repository", "Repository already exists")
	}

	s, ok := ws.Packages.StructTypeByID(cmd.Struct)
	if !ok {
		errGrp.Add("Struct", "Struct not found")
		return nil, errGrp.Error()
	}

	if v := slices.Collect(s.Fields.PrimaryKeys()); len(v) != 1 {
		errGrp.Add("Struct", fmt.Sprintf("Struct must have exactly one primary key field, found %d", len(v)))
		return nil, errGrp.Error()
	}

	if err := errGrp.Error(); err != nil {
		return nil, err
	}

	return []WorkspaceEvent{RepositoryAssigned{
		Workspace:   cmd.Workspace,
		Struct:      cmd.Struct,
		Repository:  cmd.Repository,
		Description: cmd.Description,
	}}, nil
}
