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

type FormAlignmentUpdated struct {
	Workspace WorkspaceID  `json:"workspace"`
	Form      FormID       `json:"form"`
	ID        ViewID       `json:"id"`
	Alignment ui.Alignment `json:"alignment"`
}

func (evt FormAlignmentUpdated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormAlignmentUpdated) Discriminator() evs.Discriminator {
	return "FormAlignmentUpdated"
}

func (evt FormAlignmentUpdated) Evolve(ctx context.Context, ws *Workspace) error {
	form, ok := ws.Forms.ByID(evt.Form)
	if !ok {
		return fmt.Errorf("form %s not found", evt.Form)
	}

	v, ok := FindElementByID(form.Root, evt.ID)
	if !ok {
		return fmt.Errorf("element %s not found", evt.ID)
	}

	if v, ok := v.(Alignable); ok {
		v.SetAlignment(evt.Alignment)
	}

	return nil
}
