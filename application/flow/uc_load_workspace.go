// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewLoadWorkspace(repoName string, replay evs.ReplayWithIndex[WorkspaceID, WorkspaceEvent]) LoadWorkspace {
	return func(subject auth.Subject, id WorkspaceID) (option.Opt[*Workspace], error) {
		if err := subject.AuditResource(repoName, string(id), PermFindWorkspaces); err != nil {
			return option.None[*Workspace](), err
		}

		var ws Workspace
		if err := replay(user.SU(), id, ws.ApplyEnvelope, evs.ReplayOptions{}); err != nil {
			return option.None[*Workspace](), err
		}

		return option.Some(&ws), nil
	}
}
