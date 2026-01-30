// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xerrors"
)

type UpdateFormEnableExpr struct {
	Workspace  WorkspaceID
	Form       FormID
	ID         ViewID
	Expression Expression
}

func (cmd UpdateFormEnableExpr) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd UpdateFormEnableExpr) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	v, ok := GetView(ws, cmd.Form, cmd.ID)
	if !ok {
		errGrp.Add("ID", "View not found")
		return nil, errGrp.Error()
	}

	if _, ok := v.(Enabler); !ok {
		errGrp.Add("ID", "View is not Enabler")
		return nil, errGrp.Error()
	}

	return []WorkspaceEvent{FormEnableExprUpdated{
		Workspace:  cmd.Workspace,
		Form:       cmd.Form,
		ID:         cmd.ID,
		EnableExpr: cmd.Expression,
	}}, nil
}
