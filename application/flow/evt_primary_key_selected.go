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

type PrimaryKeySelected struct {
	Workspace WorkspaceID `json:"workspace"`
	Struct    TypeID      `json:"struct"`
	Field     FieldID     `json:"field"`
}

func (evt PrimaryKeySelected) Evolve(ctx context.Context, ws *Workspace) error {
	st, ok := ws.Packages.StructTypeByID(evt.Struct)
	if !ok {
		return fmt.Errorf("struct %s not found", evt.Struct)
	}

	for f := range st.Fields.All() {
		if f, ok := f.(PKField); ok {
			if f.Identity() == evt.Field {
				f.SetPrimaryKey(true)
			} else {
				// keep it, it is defined like this
				f.SetPrimaryKey(false)
			}
		}
	}

	return nil
}

func (evt PrimaryKeySelected) Discriminator() evs.Discriminator {
	return "PrimaryKeySelected"
}

func (evt PrimaryKeySelected) WorkspaceID() WorkspaceID {
	return evt.Workspace
}
