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

type FormTextFieldAdded struct {
	Workspace      WorkspaceID `json:"workspace"`
	Form           FormID      `json:"form"`
	Parent         ViewID      `json:"parent"`
	After          ViewID      `json:"after,omitempty"` // optional, if empty add as first element (even if not empty), otherwise after (usually below or right of)
	ID             ViewID      `json:"id"`
	Label          string      `json:"label"`
	SupportingText string      `json:"supportingText"`
	Lines          int         `json:"lines"`
	Field          FieldID     `json:"field"`
}

func (evt FormTextFieldAdded) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormTextFieldAdded) Discriminator() evs.Discriminator {
	return "FormTextFieldAdded"
}

func (evt FormTextFieldAdded) Evolve(ctx context.Context, ws *Workspace) error {
	form, ok := ws.Forms.ByID(evt.Form)
	if !ok {
		return fmt.Errorf("form %s not found", evt.Form)
	}

	structType := form.Type()

	parent, ok := GetViewGroup(ws, evt.Form, evt.Parent)
	if !ok {
		return fmt.Errorf("parent %s not found", evt.Form)
	}

	v := NewFormTextField(evt.ID, structType, evt.Field)
	v.SetLabel(evt.Label)
	v.SetSupportingText(evt.SupportingText)
	v.SetLines(evt.Lines)
	v.SetFrame(ui.Frame{}.FullWidth())

	parent.Insert(v, evt.After)
	return nil
}
