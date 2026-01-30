// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"slices"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xerrors"
	"go.wdy.de/nago/presentation/ui"
)

type UpdateButtonStyle struct {
	Workspace WorkspaceID
	Form      FormID
	ID        ViewID
	Style     ui.ButtonStyle
}

func (cmd UpdateButtonStyle) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd UpdateButtonStyle) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	v, ok := GetView(ws, cmd.Form, cmd.ID)
	if !ok {
		errGrp.Add("ID", "View not found")
		return nil, errGrp.Error()
	}

	if _, ok := v.(*FormButton); !ok {
		errGrp.Add("ID", "View is not a FormButton")
		return nil, errGrp.Error()
	}

	if !slices.Contains(ui.ButtonStyles(), cmd.Style) {
		errGrp.Add("Style", "Style not valid")
		return nil, errGrp.Error()
	}

	return []WorkspaceEvent{ButtonStyleUpdated{
		Workspace: cmd.Workspace,
		Form:      cmd.Form,
		ID:        cmd.ID,
		Style:     cmd.Style,
	}}, nil
}
