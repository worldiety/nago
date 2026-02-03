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

type AddCheckboxCmd struct {
	Workspace      WorkspaceID `visible:"false"`
	Form           FormID      `visible:"false"`
	Parent         ViewID      `visible:"false"`
	After          ViewID      `visible:"false"`
	Label          string
	SupportingText string  `lines:"3"`
	Field          FieldID `source:"nago.flow.fields" dialogOptions:"larger"`
}

func (cmd AddCheckboxCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd AddCheckboxCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	form, ok := ws.Forms.ByID(cmd.Form)
	if !ok {
		errGrp.Add("Form", "Form not found")
		return nil, errGrp.Error()
	}

	structType, ok := ws.Packages.StructTypeByID(form.Type())
	if !ok {
		errGrp.Add("Form", "Repository type not found")
		return nil, errGrp.Error()
	}

	_, ok = structType.Fields.ByID(cmd.Field)
	if !ok {
		errGrp.Add("Field", "Field not found")
		return nil, errGrp.Error()
	}

	parent, ok := GetViewGroup(ws, cmd.Form, cmd.Parent)
	if !ok {
		errGrp.Add("Form", "Form not found")
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

	return []WorkspaceEvent{FormCheckboxAdded{
		Workspace:      cmd.Workspace,
		Form:           cmd.Form,
		Parent:         parent.Identity(),
		After:          cmd.After,
		ID:             id,
		Label:          cmd.Label,
		Field:          cmd.Field,
		SupportingText: cmd.SupportingText,
	}}, nil
}
