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

type UpdateFormVisibleExpr struct {
	Workspace  WorkspaceID
	Form       FormID
	ID         ViewID
	Expression Expression
}

func (cmd UpdateFormVisibleExpr) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd UpdateFormVisibleExpr) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	v, ok := GetView(ws, cmd.Form, cmd.ID)
	if !ok {
		errGrp.Add("ID", "View not found")
		return nil, errGrp.Error()
	}

	// TODO the API of expr is not clear to me and it works differently than yaegi. It seems that checking and parsing is tied to the given actual types which we don't have at this point
	// We use this, because it also as zero deps but does not depend on any special go versions which yaegi does.
	// Even though it is backed by traefic, I'm not feeling well, if we cannot move forward, as Go progresses.
	v.SetVisibleExpr(cmd.Expression)

	return []WorkspaceEvent{FormVisibleExprUpdated{
		Workspace:  cmd.Workspace,
		Form:       cmd.Form,
		ID:         cmd.ID,
		Expression: cmd.Expression,
	}}, nil
}
