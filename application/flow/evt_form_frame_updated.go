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

type FormFrameUpdated struct {
	Workspace WorkspaceID `json:"workspace"`
	Form      FormID      `json:"form"`
	ID        ViewID      `json:"id"`
	Frame     ui.Frame    `json:"frame"`
}

func (evt FormFrameUpdated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormFrameUpdated) Discriminator() evs.Discriminator {
	return "FormFrameUpdated"
}

func (evt FormFrameUpdated) Evolve(ctx context.Context, ws *Workspace) error {
	form, ok := ws.Forms.ByID(evt.Form)
	if !ok {
		return fmt.Errorf("form %s not found", evt.Form)
	}

	v, ok := FindElementByID(form.Root, evt.ID)
	if !ok {
		return fmt.Errorf("element %s not found", evt.ID)
	}

	if v, ok := v.(Frameable); ok {
		v.SetFrame(evt.Frame)
	}

	return nil
}
