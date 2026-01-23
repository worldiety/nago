// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xerrors"
)

type CreatePackageCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	Path        ImportPath
	Name        Ident
	Description string `lines:"3"`
}

func (cmd CreatePackageCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd CreatePackageCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	pid := data.RandIdent[PackageID]()
	if _, ok := ws.Packages.ByID(pid); ok {
		return nil, fmt.Errorf("package %s collision", pid)
	}

	var errGrp xerrors.FieldBuilder
	if err := cmd.Path.Validate(); err != nil {
		errGrp.Add("Path", "Invalid Path: "+err.Error())
	}

	if _, ok := ws.Packages.ByImportPath(cmd.Path); ok {
		errGrp.Add("Path", "Package already exists at path")
	}

	if err := cmd.Name.Validate(); err != nil {
		errGrp.Add("Name", "Invalid Name: "+err.Error())
	}

	if !cmd.Name.IsPrivate() {
		errGrp.Add("Name", "Name must start with a lowercase letter")
	}

	if err := errGrp.Error(); err != nil {
		return nil, err
	}

	return []WorkspaceEvent{PackageCreated{
		Workspace:   ws.ID,
		Package:     pid,
		Path:        cmd.Path,
		Name:        cmd.Name,
		Description: cmd.Description,
	}}, nil
}

func (cmd CreatePackageCmd) cmd() {}
