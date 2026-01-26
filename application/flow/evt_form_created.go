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
)

type FormID string

type FormCreated struct {
	Workspace   WorkspaceID  `json:"workspace"`
	Repository  RepositoryID `json:"repository"`
	ID          FormID       `json:"id"`
	Name        Ident        `json:"name"`
	Description string       `json:"description"`
}

func (evt FormCreated) Discriminator() evs.Discriminator {
	return "FormCreated"
}

func (evt FormCreated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormCreated) Evolve(ctx context.Context, ws *Workspace) error {
	vstack := NewFormVStack(ViewID(evt.ID + "-root"))
	form := NewForm(evt.ID, evt.Name, vstack)

	ws.Forms.AddForm(form)
	return nil
}
