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

type StringTypeCreated struct {
	Workspace   WorkspaceID `json:"workspace,omitempty"`
	ID          TypeID      `json:"id,omitempty"`
	Name        Ident       `json:"name,omitempty"`
	Package     PackageID   `json:"package,omitempty"`
	Description string      `json:"description,omitempty"`
}

func (evt StringTypeCreated) Discriminator() evs.Discriminator {
	return "StringTypeCreated"
}

func (evt StringTypeCreated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt StringTypeCreated) event() {}

func (evt StringTypeCreated) Evolve(ws *Workspace) error {
	pkg, ok := ws.Packages.ByID(evt.Package)
	if !ok {
		return fmt.Errorf("package %s not found", evt.Package)
	}

	t := NewStringType(pkg.ID, evt.ID, evt.Name)
	t.SetDescription(evt.Description)
	if !pkg.Types.AddType(t) {
		return fmt.Errorf("type %s already exists", evt.Name)
	}

	return nil
}
