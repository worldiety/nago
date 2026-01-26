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

type BoolFieldAppended struct {
	Workspace   WorkspaceID `json:"workspace"`
	Struct      TypeID      `json:"struct"`
	Name        Ident       `json:"name"`
	ID          FieldID     `json:"id"`
	Description string      `json:"description"`
}

func (evt BoolFieldAppended) Discriminator() evs.Discriminator {
	return "BoolFieldAppended"
}

func (evt BoolFieldAppended) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt BoolFieldAppended) Evolve(ctx context.Context, ws *Workspace) error {
	var errGrp xerrors.FieldBuilder
	st, ok := ws.Packages.StructTypeByID(evt.Struct)
	if !ok {
		errGrp.Add("Struct", "Struct not found")
		return errGrp.Error()
	}

	f := NewBoolField(st.ID, evt.ID, evt.Name)
	f.SetDescription(evt.Description)
	st.Fields.Add(f)
	return nil
}
