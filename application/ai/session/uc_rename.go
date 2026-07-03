// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"
	"sync"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

// NewRename returns a [Rename] use case that updates the title of a session. Requires PermRename globally or
// as an instance grant on the session.
func NewRename(mutex *sync.Mutex, repo Repository) Rename {
	return func(subject auth.Subject, id ID, title string) error {
		mutex.Lock()
		defer mutex.Unlock()

		optSession, err := repo.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot load session: %w", err)
		}

		if optSession.IsNone() {
			return fmt.Errorf("session %q does not exist", id)
		}

		if err := subject.AuditResource(Namespace, rebacInstance(id), PermRename); err != nil {
			return fmt.Errorf("session %q does not exist", id)
		}

		session := optSession.Unwrap()
		session.Title = title
		session.UpdatedAt = xtime.Now()

		if err := repo.Save(session); err != nil {
			return fmt.Errorf("cannot persist session: %w", err)
		}

		return nil
	}
}
