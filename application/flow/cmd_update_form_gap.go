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

type UpdateFormGap struct {
	Workspace WorkspaceID
	Form      FormID
	ID        ViewID
	Gap       ui.Length
}

func (cmd UpdateFormGap) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd UpdateFormGap) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	v, ok := GetViewGroup(ws, cmd.Form, cmd.ID)
	if !ok {
		errGrp.Add("ID", "View not found")
		return nil, errGrp.Error()
	}

	if _, ok := v.(Gapable); !ok {
		errGrp.Add("ID", "View has not gap property")
		return nil, errGrp.Error()
	}

	return []WorkspaceEvent{FormGapUpdated{
		Workspace: cmd.Workspace,
		Form:      cmd.Form,
		ID:        cmd.ID,
		Gap:       cmd.Gap,
	}}, nil
}
