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

type FormButtonAdded struct {
	Workspace WorkspaceID    `json:"workspace"`
	Form      FormID         `json:"form"`
	Parent    ViewID         `json:"parent"`
	After     ViewID         `json:"after,omitempty"` // optional, if empty add as first element (even if not empty), otherwise after (usually below or right of)
	ID        ViewID         `json:"id"`
	Title     string         `json:"title"`
	Style     ui.ButtonStyle `json:"style"`
}

func (evt FormButtonAdded) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormButtonAdded) Discriminator() evs.Discriminator {
	return "FormButtonAdded"
}

func (evt FormButtonAdded) Evolve(ctx context.Context, ws *Workspace) error {
	parent, ok := GetViewGroup(ws, evt.Form, evt.Parent)
	if !ok {
		return fmt.Errorf("parent %s not found", evt.Form)
	}

	btn := NewFormButton(evt.ID, evt.Title, evt.Style)
	parent.Insert(btn, evt.After)
	return nil
}
