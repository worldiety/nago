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

func NewCreatePackage(hnd handleCmd[PackageCreated]) CreatePackage {
	return func(subject auth.Subject, cmd CreatePackageCmd) (PackageCreated, error) {
		return hnd(subject, cmd.Workspace, func(ws *Workspace) (PackageCreated, error) {
			var zero PackageCreated

			pid := data.RandIdent[PackageID]()
			if _, ok := ws.packages[pid]; ok {
				return zero, fmt.Errorf("package %s already exists", pid)
			}

			if err := cmd.Path.Validate(); err != nil {
				return zero, xerrors.WithFields("Failure", "Path", "Invalid Path: "+err.Error())
			}

			if _, ok := ws.packageByPath(cmd.Path); ok {
				return zero, xerrors.WithFields("Failure", "Path", "Package already exists at path")
			}

			if err := cmd.Name.Validate(); err != nil {
				return zero, xerrors.WithFields("Failure", "Name", "Invalid Name: "+err.Error())
			}

			if !cmd.Name.IsPrivate() {
				return zero, xerrors.WithFields("Failure", "Name", "Name must start with a lowercase letter")
			}

			return PackageCreated{
				Workspace:   ws.id,
				Package:     pid,
				Path:        cmd.Path,
				Name:        cmd.Name,
				Description: cmd.Description,
			}, nil
		})
	}
}
