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
	"go.wdy.de/nago/pkg/xslices"
)

type AddFormTextCmd struct {
	Workspace WorkspaceID   `visible:"false"`
	Form      FormID        `visible:"false"`
	Parent    ViewID        `visible:"false"`
	After     ViewID        `visible:"false"`
	Renderer  RendererID    `visible:"false"`
	Value     string        `label:"nago.common.label.value" lines:"5"`
	Style     FormTextStyle `values:"[\"text\", \"h1\", \"h2\", \"h3\", \"markdown\"]"`
}

func (cmd AddFormTextCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd AddFormTextCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	parent, ok := GetViewGroup(ws, cmd.Form, cmd.Parent)
	if !ok {
		errGrp.Add("Form", "Form not found")
		return nil, errGrp.Error()
	}

	if cmd.Renderer == "" {
		errGrp.Add("Renderer", "Renderer not set")
	}

	if !xslices.Contains(FormTextValues, cmd.Style) {
		errGrp.Add("Style", "Style not valid")
	}

	if cmd.After != "" {
		if _, ok := FindElementByID(parent, cmd.After); ok {
			errGrp.Add("After", "After element not found")
		}
	}

	if err := errGrp.Error(); err != nil {
		return nil, err
	}

	id := data.RandIdent[ViewID]()

	return []WorkspaceEvent{FormTextAdded{
		Workspace: cmd.Workspace,
		Form:      cmd.Form,
		Parent:    parent.Identity(),
		After:     cmd.After,
		ID:        id,
		Value:     cmd.Value,
		Style:     cmd.Style,
	}}, nil
}
