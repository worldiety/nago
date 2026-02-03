// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xerrors"
)

type AddFormRefCmd struct {
	Workspace WorkspaceID `visible:"false"`
	Form      FormID      `visible:"false"`
	Parent    ViewID      `visible:"false"`
	After     ViewID      `visible:"false"`
	Ref       FormID      `source:"nago.flow.forms"`
}

func (cmd AddFormRefCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd AddFormRefCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	parent, ok := GetViewGroup(ws, cmd.Form, cmd.Parent)
	if !ok {
		errGrp.Add("Form", "Form not found")
		return nil, errGrp.Error()
	}

	if _, ok := ws.Forms.ByID(cmd.Ref); !ok {
		errGrp.Add("Ref", "Ref Form not found")
		return nil, errGrp.Error()
	}

	if cmd.After != "" {
		if _, ok := FindElementByID(parent, cmd.After); !ok {
			errGrp.Add("After", "After element not found")
		}
	}

	if err := errGrp.Error(); err != nil {
		return nil, err
	}

	id := data.RandIdent[ViewID]()

	return []WorkspaceEvent{FormRefAdded{
		Workspace: cmd.Workspace,
		Form:      cmd.Form,
		Parent:    parent.Identity(),
		After:     cmd.After,
		ID:        id,
		Ref:       cmd.Ref,
	}}, nil
}
