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

type UpdateFormBorder struct {
	Workspace WorkspaceID
	Form      FormID
	ID        ViewID
	Border    ui.Border
}

func (cmd UpdateFormBorder) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd UpdateFormBorder) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	v, ok := GetView(ws, cmd.Form, cmd.ID)
	if !ok {
		errGrp.Add("ID", "View not found")
		return nil, errGrp.Error()
	}

	if _, ok := v.(Borderable); !ok {
		errGrp.Add("ID", "View has not border property")
		return nil, errGrp.Error()
	}

	return []WorkspaceEvent{FormBorderUpdated{
		Workspace: cmd.Workspace,
		Form:      cmd.Form,
		ID:        cmd.ID,
		Border:    cmd.Border,
	}}, nil
}
