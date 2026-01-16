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
	"go.wdy.de/nago/pkg/xerrors"
)

func NewCreateWorkspace(storeEvent evs.Store[WorkspaceEvent]) CreateWorkspace {
	return func(subject auth.Subject, cmd CreateWorkspaceCmd) (WorkspaceCreated, error) {
		var zero WorkspaceCreated
		if err := subject.Audit(PermCreateWorkspace); err != nil {
			return zero, err
		}

		if err := cmd.Name.Validate(); err != nil {
			return zero, xerrors.WithFields("Validation", "Name", err.Error())
		}

		if cmd.Name.IsPublic() {
			return zero, xerrors.WithFields("Validation", "Name", "Name must not start with an uppercase letter")
		}

		id := data.RandIdent[WorkspaceID]()
		evt := WorkspaceCreated{
			Workspace:   id,
			Name:        cmd.Name,
			Description: cmd.Description,
		}

		_, err := storeEvent(user.SU(), evt, evs.StoreOptions{
			CreatedBy: subject.ID(),
		})

		if err != nil {
			return zero, err
		}

		return evt, nil
	}
}
