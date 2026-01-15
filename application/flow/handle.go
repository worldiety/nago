// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"os"

	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

type handleCmd[T WorkspaceEvent] func(subject auth.Subject, id WorkspaceID, fn func(ws *Workspace) (T, error)) (T, error)

func newHandleCmd[T WorkspaceEvent](repoName string, load LoadWorkspace, store evs.Store[WorkspaceEvent]) handleCmd[T] {

	return func(subject auth.Subject, id WorkspaceID, fn func(ws *Workspace) (T, error)) (T, error) {
		var zero T
		if err := subject.AuditResource(repoName, string(id), PermUpdateWorkspace); err != nil {
			return zero, err
		}

		optWS, err := load(user.SU(), id)
		if err != nil {
			return zero, err
		}

		if optWS.IsNone() {
			return zero, fmt.Errorf("workspace %s not found: %w", id, os.ErrNotExist)
		}

		ws := optWS.Unwrap()
		ws.mutex.Lock()
		evt, err := fn(ws)
		ws.mutex.Unlock()
		if err != nil {
			return zero, err
		}

		if _, err := store(user.SU(), evt, evs.StoreOptions{
			CreatedBy: subject.ID(),
		}); err != nil {
			return zero, err
		}

		if err := ws.apply(evt); err != nil {
			return zero, err
		}

		return evt, nil
	}
}
