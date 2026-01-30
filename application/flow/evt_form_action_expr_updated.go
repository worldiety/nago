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

type FormActionExprUpdated struct {
	Workspace WorkspaceID  `json:"workspace"`
	Form      FormID       `json:"form"`
	ID        ViewID       `json:"id"`
	Action    []Expression `json:"action"`
}

func (evt FormActionExprUpdated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormActionExprUpdated) Discriminator() evs.Discriminator {
	return "FormActionExprUpdated"
}

func (evt FormActionExprUpdated) Evolve(ctx context.Context, ws *Workspace) error {
	v, ok := GetView(ws, evt.Form, evt.ID)
	if !ok {
		return fmt.Errorf("view %s not found", evt.ID)
	}

	ac, ok := v.(Actionable)
	if !ok {
		return fmt.Errorf("view %s is not actionable", evt.ID)
	}

	ac.SetActionExpr(evt.Action...)

	return nil
}
