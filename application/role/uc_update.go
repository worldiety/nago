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

		if err := repo.Save(role); err != nil {
			return err
		}

		bus.Publish(Updated{Role: role.ID})

		return nil
	}
}
