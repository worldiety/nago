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

func NewDeleteWorkspace(handler *evs.Handler[*Workspace, WorkspaceEvent, WorkspaceID]) DeleteWorkspace {
	return func(subject auth.Subject, id WorkspaceID) error {
		if err := subject.Audit(PermDeleteWorkspace); err != nil {
			return err
		}

		return handler.Delete(subject, id)
	}
}
