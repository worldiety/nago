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
	"go.wdy.de/nago/presentation/ui"
)

type UpdateFormFrame struct {
	Workspace WorkspaceID
	Form      FormID
	ID        ViewID
	Frame     ui.Frame
}

func (cmd UpdateFormFrame) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd UpdateFormFrame) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	v, ok := GetView(ws, cmd.Form, cmd.ID)
	if !ok {
		errGrp.Add("ID", "View not found")
		return nil, errGrp.Error()
	}

	if _, ok := v.(Frameable); !ok {
		errGrp.Add("ID", "View has not frame property")
		return nil, errGrp.Error()
	}

	f := cmd.Frame
	f.Width = FixLength(f.Width)
	f.Height = FixLength(f.Height)
	f.MinWidth = FixLength(f.MinWidth)
	f.MinHeight = FixLength(f.MinHeight)
	f.MaxWidth = FixLength(f.MaxWidth)
	f.MaxHeight = FixLength(f.MaxHeight)

	return []WorkspaceEvent{FormFrameUpdated{
		Workspace: cmd.Workspace,
		Form:      cmd.Form,
		ID:        cmd.ID,
		Frame:     f,
	}}, nil
}
