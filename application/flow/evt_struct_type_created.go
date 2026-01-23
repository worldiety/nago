// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"

	"go.wdy.de/nago/application/evs"
)

type StructTypeCreated struct {
	Workspace   WorkspaceID `json:"workspace,omitempty"`
	ID          TypeID      `json:"id,omitempty"`
	Name        Ident       `json:"name,omitempty"`
	Package     PackageID   `json:"package,omitempty"`
	Description string      `json:"description,omitempty"`
}

func (evt StructTypeCreated) Discriminator() evs.Discriminator {
	return "StructTypeCreated"
}

func (evt StructTypeCreated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt StructTypeCreated) event() {}

func (evt StructTypeCreated) Evolve(ws *Workspace) error {
	pkg, _ := ws.Packages.ByID(evt.Package)
	if pkg == nil {
		return fmt.Errorf("package %s not found", evt.Package)
	}

	t := NewStructType(pkg.ID, evt.ID, evt.Name)
	t.SetDescription(evt.Description)
	pkg.Types.AddType(t)
	return nil
}
