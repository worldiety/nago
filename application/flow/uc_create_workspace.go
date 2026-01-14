// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewCreateWorkspace(storeEvent evs.Store[WorkspaceEvent]) CreateWorkspace {
	return func(subject auth.Subject, cmd CreateWorkspaceCmd) (WorkspaceID, error) {
		if err := subject.Audit(PermCreateWorkspace); err != nil {
			return "", err
		}

		id := data.RandIdent[WorkspaceID]()
		var ws Workspace
		evt, err := storeEvent(user.SU(), WorkspaceCreated{
			Workspace:   id,
			Name:        cmd.Name,
			Description: cmd.Description,
		}, evs.StoreOptions{
			CreatedBy: subject.ID(),
		})
		
		if err != nil {
			return id, err
		}

		return id, ws.ApplyEnvelope(evt)
	}
}
