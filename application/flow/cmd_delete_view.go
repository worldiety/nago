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

type DeleteViewCmd struct {
	Workspace WorkspaceID `visible:"false"`
	Form      FormID      `visible:"false"`
	View      ViewID      `visible:"false"`
}

func (cmd DeleteViewCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd DeleteViewCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	form, ok := ws.Forms.ByID(cmd.Form)
	if !ok {
		errGrp.Add("Form", "Form not found")
		return nil, errGrp.Error()
	}

	if _, ok := FindElementByID(form.Root, cmd.View); !ok {
		errGrp.Add("View", "View not found")
		return nil, errGrp.Error()
	}

	return []WorkspaceEvent{FormViewDeleted{
		Workspace: cmd.Workspace,
		Form:      cmd.Form,
		ID:        cmd.View,
	}}, nil
}
