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

type UpdateFormAlignment struct {
	Workspace WorkspaceID
	Form      FormID
	ID        ViewID
	Alignment ui.Alignment
}

func (cmd UpdateFormAlignment) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd UpdateFormAlignment) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	var errGrp xerrors.FieldBuilder

	v, ok := GetViewGroup(ws, cmd.Form, cmd.ID)
	if !ok {
		errGrp.Add("ID", "View not found")
		return nil, errGrp.Error()
	}

	if _, ok := v.(Alignable); !ok {
		errGrp.Add("ID", "View is not alignable")
		return nil, errGrp.Error()
	}

	if !slices.Contains(ui.Alignments(), cmd.Alignment) {
		errGrp.Add("Alignment", "Alignment not valid")
		return nil, errGrp.Error()
	}

	return []WorkspaceEvent{FormAlignmentUpdated{
		Workspace: cmd.Workspace,
		Form:      cmd.Form,
		ID:        cmd.ID,
		Alignment: cmd.Alignment,
	}}, nil
}
