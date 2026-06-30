// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

// deleteWorkspaceCmd produces a [WorkspaceDeleted] event whose Evolve marks the
// aggregate deleted, causing the handler to drop it from its in-memory set.
type deleteWorkspaceCmd struct {
	id WorkspaceID
}

func (cmd deleteWorkspaceCmd) Decide(subject auth.Subject, ws *Workspace) ([]WorkspaceEvent, error) {
	return []WorkspaceEvent{WorkspaceDeleted{Workspace: cmd.id}}, nil
}

func NewDeleteWorkspace(handler *evs.Handler[*Workspace, WorkspaceEvent, WorkspaceID]) DeleteWorkspace {
	return func(subject auth.Subject, id WorkspaceID) error {
		if err := subject.Audit(PermDeleteWorkspace); err != nil {
			return err
		}

		return handler.Delete(subject, id, deleteWorkspaceCmd{id: id})
	}
}
