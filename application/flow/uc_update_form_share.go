// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"slices"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewUpdateFormShare(idx *shareIndex, load LoadWorkspace) UpdateFormShare {
	return func(subject auth.Subject, workspaceID WorkspaceID, formID FormID, opts ShareFormOptions) (FormShareID, error) {
		var shareId FormShareID
		err := idx.byForm(workspaceID, formID, func(ws *Workspace, form *Form) error {

			share, ok := idx.reverseLookup[formID]
			if !ok {
				share = FormShare{
					ID:        data.RandIdent[FormShareID](),
					Workspace: workspaceID,
					Form:      formID,
				}
			}

			share.AllowUnauthenticated = opts.AllowUnauthenticated
			share.AllowedUsers = slices.Clone(opts.AllowedUsers)

			isAllowed := subject.HasPermission(PermFindWorkspaces) || ws.IsOwner(subject.ID())
			if !isAllowed {
				return user.PermissionDeniedError("workspace forbids access")
			}

			if err := idx.repo.Save(share); err != nil {
				return fmt.Errorf("failed to save form share: %w", err)
			}

			idx.reverseLookup[formID] = share
			idx.lookup[share.Identity()] = share
			shareId = share.ID

			return nil
		})

		return shareId, err
	}
}
