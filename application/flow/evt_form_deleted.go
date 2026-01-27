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

type FormDeleted struct {
	Workspace WorkspaceID `json:"workspace"`
	ID        FormID      `json:"id"`
}

func (evt FormDeleted) Discriminator() evs.Discriminator {
	return "FormDeleted"
}

func (evt FormDeleted) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormDeleted) Evolve(ctx context.Context, ws *Workspace) error {
	ws.Forms.Remove(evt.ID)
	return nil
}
