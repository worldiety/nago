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
	"go.wdy.de/nago/presentation/ui"
)

type FormBackgroundColorUpdated struct {
	Workspace WorkspaceID `json:"workspace"`
	Form      FormID      `json:"form"`
	ID        ViewID      `json:"id"`
	Color     ui.Color    `json:"color"`
}

func (evt FormBackgroundColorUpdated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormBackgroundColorUpdated) Discriminator() evs.Discriminator {
	return "FormBackgroundColorUpdated"
}

func (evt FormBackgroundColorUpdated) Evolve(ctx context.Context, ws *Workspace) error {
	form, ok := ws.Forms.ByID(evt.Form)
	if !ok {
		return fmt.Errorf("form %s not found", evt.Form)
	}

	v, ok := FindElementByID(form.Root, evt.ID)
	if !ok {
		return fmt.Errorf("element %s not found", evt.ID)
	}

	if v, ok := v.(Backgroundable); ok {
		v.SetBackgroundColor(evt.Color)
	}

	return nil
}
