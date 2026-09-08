// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package role

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
	"strings"
	"sync"
)

func NewUpdate(mutex *sync.Mutex, repo Repository, bus events.Bus) Update {
	return func(subject permission.Auditable, role Role) error {
		if err := subject.Audit(PermUpdate); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if strings.TrimSpace(string(role.ID)) == "" {
			return std.NewLocalizedError("Ungültige EID", "Eine leere Gruppen EID ist nicht zulässig.")
		}

		// The System flag is owned by the declaring module (see [Configurator.DeclareSystemRole]), never by
		// the caller of Update: otherwise the editing UI, which round-trips the whole aggregate, could turn a
		// protected role into an ordinary one and delete it right afterwards. Name and description stay
		// editable, so an operator may still clarify the wording for their organisation.
		optExisting, err := repo.FindByID(role.ID)
		if err != nil {
			return err
		}

		if optExisting.IsSome() {
			role.System = optExisting.Unwrap().System
		} else {
			role.System = false
		}

		if err := repo.Save(role); err != nil {
			return err
		}

		bus.Publish(Updated{Role: role.ID})

		return nil
	}
}
