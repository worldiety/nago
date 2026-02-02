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

type FormHLineAdded struct {
	Workspace WorkspaceID `json:"workspace"`
	Form      FormID      `json:"form"`
	Parent    ViewID      `json:"parent"`
	After     ViewID      `json:"after,omitempty"` // optional, if empty add as first element (even if not empty), otherwise after (usually below or right of)
	ID        ViewID      `json:"id"`
}

func (evt FormHLineAdded) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormHLineAdded) Discriminator() evs.Discriminator {
	return "FormHLineAdded"
}

func (evt FormHLineAdded) Evolve(ctx context.Context, ws *Workspace) error {
	form, ok := GetViewGroup(ws, evt.Form, evt.Parent)
	if !ok {
		return fmt.Errorf("form %s not found", evt.Form)
	}

	v := NewFormHLine(evt.ID)
	fr := v.Frame()
	fr.Width = ui.Full
	v.SetFrame(fr)

	pd := v.Padding()
	pd = pd.Vertical(ui.L16)
	v.SetPadding(pd)

	form.Insert(v, evt.After)
	return nil
}
