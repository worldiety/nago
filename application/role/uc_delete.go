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
	"sync"
)

func NewDelete(mutex *sync.Mutex, repo Repository, bus events.Bus) Delete {
	return func(subject permission.Auditable, id ID) error {
		if err := subject.Audit(PermDelete); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if err := repo.DeleteByID(id); err != nil {
			return err
		}

		bus.Publish(Deleted{Role: id})
		return nil
	}
}
