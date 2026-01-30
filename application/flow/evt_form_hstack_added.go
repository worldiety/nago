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

type FormHStackAdded struct {
	Workspace WorkspaceID `json:"workspace"`
	Form      FormID      `json:"form"`
	Parent    ViewID      `json:"parent"`
	After     ViewID      `json:"after,omitempty"` // optional, if empty add as first element (even if not empty), otherwise after (usually below or right of)
	ID        ViewID      `json:"id"`
}

func (evt FormHStackAdded) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormHStackAdded) Discriminator() evs.Discriminator {
	return "FormHStackAdded"
}

func (evt FormHStackAdded) Evolve(ctx context.Context, ws *Workspace) error {
	form, ok := GetViewGroup(ws, evt.Form, evt.Parent)
	if !ok {
		return fmt.Errorf("form %s not found", evt.Form)
	}

	form.Insert(NewFormHStack(evt.ID), evt.After)
	return nil
}
