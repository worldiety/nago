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

type WorkspaceDeleted struct {
	Workspace WorkspaceID `json:"workspace"`
}

func (evt WorkspaceDeleted) Discriminator() evs.Discriminator {
	return "WorkspaceDeleted"
}

func (evt WorkspaceDeleted) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt WorkspaceDeleted) Evolve(ctx context.Context, ws *Workspace) error {
	ws.deleted = true
	return nil
}
