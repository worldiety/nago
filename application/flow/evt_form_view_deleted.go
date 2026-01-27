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

type FormViewDeleted struct {
	Workspace WorkspaceID `json:"workspace"`
	Form      FormID      `json:"form"`
	ID        ViewID      `json:"id"`
}

func (evt FormViewDeleted) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormViewDeleted) Discriminator() evs.Discriminator {
	return "FormViewDeleted"
}

func (evt FormViewDeleted) Evolve(ctx context.Context, ws *Workspace) error {
	form, ok := ws.Forms.ByID(evt.Form)
	if !ok {
		return nil // fine, its gone
	}

	if form.Root != nil && form.Root.Identity() == evt.ID {
		form.Root = nil
		return nil
	}

	DeleteElementByID(form.Root, evt.ID)
	return nil
}
