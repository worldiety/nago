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
	"go.wdy.de/nago/pkg/cloner"
)

type WorkspaceID string

type WorkspaceCreated struct {
	Workspace   WorkspaceID `json:"workspace,omitempty"`
	Name        Ident       `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
}

func (evt WorkspaceCreated) Evolve(ctx context.Context, ws *Workspace) error {
	ws.ID = evt.Workspace
	ws.Packages = NewPackages()
	ws.Repositories = NewRepositories()
	ws.Forms = NewForms()
	ws.Extensions = map[string]cloner.Cloneable{}
	ws.Name = evt.Name
	ws.Description = evt.Description
	return nil
}

func (evt WorkspaceCreated) Discriminator() evs.Discriminator {
	return "WorkspaceCreated"
}

func (evt WorkspaceCreated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt WorkspaceCreated) event() {}
