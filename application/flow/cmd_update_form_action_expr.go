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

type UpdateFormActionExpr struct {
	Workspace WorkspaceID
	Form      FormID
	ID        ViewID
	Action    []Expression
}

func (cmd UpdateFormActionExpr) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd UpdateFormActionExpr) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	v, ok := GetView(ws, cmd.Form, cmd.ID)
	if !ok {
		errGrp.Add("ID", "View not found")
		return nil, errGrp.Error()
	}

	action, ok := v.(Actionable)
	if !ok {
		errGrp.Add("ID", "View is not Actionable")
		return nil, errGrp.Error()
	}
	action.SetActionExpr(cmd.Action...)

	return []WorkspaceEvent{FormActionExprUpdated{
		Workspace: cmd.Workspace,
		Form:      cmd.Form,
		ID:        cmd.ID,
		Action:    cmd.Action,
	}}, nil
}
