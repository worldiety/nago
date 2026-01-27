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

type FormCheckboxAdded struct {
	Workspace      WorkspaceID `json:"workspace"`
	Form           FormID      `json:"form"`
	Parent         ViewID      `json:"parent"`
	After          ViewID      `json:"after,omitempty"` // optional, if empty add as first element (even if not empty), otherwise after (usually below or right of)
	ID             ViewID      `json:"id"`
	Label          string      `json:"label"`
	SupportingText string      `json:"supportingText"`
	Field          FieldID     `json:"field"`
}

func (evt FormCheckboxAdded) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormCheckboxAdded) Discriminator() evs.Discriminator {
	return "FormCheckboxAdded"
}

func (evt FormCheckboxAdded) Evolve(ctx context.Context, ws *Workspace) error {
	form, ok := ws.Forms.ByID(evt.Form)
	if !ok {
		return fmt.Errorf("form %s not found", evt.Form)
	}

	structType := form.RepositoryType()

	parent, ok := GetViewGroup(ws, evt.Form, evt.Parent)
	if !ok {
		return fmt.Errorf("parent %s not found", evt.Form)
	}

	cb := NewFormCheckbox(evt.ID, structType, evt.Field)
	cb.SetSupportingText(evt.SupportingText)
	cb.SetLabel(evt.Label)
	parent.Insert(cb, evt.After)
	return nil
}
