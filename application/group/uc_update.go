// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package group

import (
	"strings"
	"sync"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
)

func NewUpdate(mutex *sync.Mutex, bus events.Bus, repo Repository) Update {
	return func(subject permission.Auditable, group Group) error {
		if err := subject.Audit(PermUpdate); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if strings.TrimSpace(string(group.ID)) == "" {
			return std.NewLocalizedError("Ungültige EID", "Eine leere Gruppen EID ist nicht zulässig.")
		}

		if err := repo.Save(group); err != nil {
			return err
		}

		bus.Publish(Updated{Group: group.ID})
		return nil
	}
}
