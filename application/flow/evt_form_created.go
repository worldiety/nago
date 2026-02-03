// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"context"

	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/pkg/xerrors"
)

type FormID string

type FormCreated struct {
	Workspace   WorkspaceID `json:"workspace"`
	Package     PackageID   `json:"package"`
	Struct      TypeID      `json:"struct"`
	ID          FormID      `json:"id"`
	Name        Ident       `json:"name"`
	Description string      `json:"description"`
}

func (evt FormCreated) Discriminator() evs.Discriminator {
	return "FormCreated"
}

func (evt FormCreated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormCreated) Evolve(ctx context.Context, ws *Workspace) error {
	var errGrp xerrors.FieldBuilder
	_, ok := ws.Packages.StructTypeByID(evt.Struct)
	if !ok {
		errGrp.Add("Struct", "Struct not found")
		return errGrp.Error()
	}

	vstack := NewFormVStack(ViewID(evt.ID + "-root"))
	form := NewForm(evt.ID, evt.Name, vstack, evt.Package, evt.Struct)

	ws.Forms.AddForm(form)
	return nil
}
