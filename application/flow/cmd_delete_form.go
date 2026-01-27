// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"go.wdy.de/nago/auth"
)

type DeleteFormCmd struct {
	Workspace WorkspaceID `visible:"false"`
	ID        FormID      `visible:"false"`
}

func (cmd DeleteFormCmd) WorkspaceID() WorkspaceID {
	return cmd.Workspace
}

func (cmd DeleteFormCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {

	return []WorkspaceEvent{FormDeleted{Workspace: cmd.Workspace, ID: cmd.ID}}, nil
}
