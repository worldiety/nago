// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewLoadWorkspace(mutex *sync.Mutex, repoName string, replay evs.ReplayWithIndex[WorkspaceID, WorkspaceEvent], cache map[WorkspaceID]*Workspace) LoadWorkspace {
	return func(subject auth.Subject, id WorkspaceID) (option.Opt[*Workspace], error) {
		if err := subject.AuditResource(repoName, string(id), PermFindWorkspaces); err != nil {
			return option.None[*Workspace](), err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if ws, ok := cache[id]; ok {
			return option.Some(ws), nil
		}

		ws := &Workspace{valid: true}
		if err := replay(user.SU(), id, ws.applyEnvelope, evs.ReplayOptions{}); err != nil {
			return option.None[*Workspace](), err
		}

		cache[id] = ws

		return option.Some(ws), nil
	}
}
