// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"context"
	"fmt"

	"go.wdy.de/nago/application/evs"
)

type StringFieldAppended struct {
	Workspace   WorkspaceID `json:"workspace"`
	Struct      TypeID      `json:"struct"`
	Name        Ident       `json:"name"`
	ID          FieldID     `json:"id"`
	Description string      `json:"description"`
}

func (evt StringFieldAppended) Discriminator() evs.Discriminator {
	return "StringFieldAppended"
}

func (evt StringFieldAppended) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt StringFieldAppended) Evolve(ctx context.Context, ws *Workspace) error {
	st, ok := ws.Packages.StructTypeByID(evt.Struct)
	if !ok {
		return fmt.Errorf("struct %s not found", evt.Struct)
	}

	t := NewStringField(st.ID, evt.ID, evt.Name)
	t.SetDescription(evt.Description)
	st.Fields.Add(t)
	return nil
}
