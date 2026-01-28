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

type FormVisibleExprUpdated struct {
	Workspace  WorkspaceID `json:"workspace"`
	Form       FormID      `json:"form"`
	ID         ViewID      `json:"id"`
	Expression Expression  `json:"expression"`
}

func (evt FormVisibleExprUpdated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt FormVisibleExprUpdated) Discriminator() evs.Discriminator {
	return "FormVisibleExprUpdated"
}

func (evt FormVisibleExprUpdated) Evolve(ctx context.Context, ws *Workspace) error {
	v, ok := GetView(ws, evt.Form, evt.ID)
	if !ok {
		return fmt.Errorf("view %s not found", evt.ID)
	}
	
	v.SetVisibleExpr(evt.Expression)

	return nil
}
