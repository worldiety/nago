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

type FormEnableExprUpdated struct {
	Workspace  WorkspaceID `json:"workspace"`
	Form       FormID      `json:"form"`
	ID         ViewID      `json:"id"`
	EnableExpr Expression  `json:"enable"`
}

func (evt FormEnableExprUpdated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormEnableExprUpdated) Discriminator() evs.Discriminator {
	return "FormEnableExprUpdated"
}

func (evt FormEnableExprUpdated) Evolve(ctx context.Context, ws *Workspace) error {
	v, ok := GetView(ws, evt.Form, evt.ID)
	if !ok {
		return fmt.Errorf("view %s not found", evt.ID)
	}

	ac, ok := v.(Enabler)
	if !ok {
		return fmt.Errorf("view %s is not actionable", evt.ID)
	}

	ac.SetEnabledExpr(evt.EnableExpr)

	return nil
}
