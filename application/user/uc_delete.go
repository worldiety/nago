// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"time"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/events"
)

func NewDelete(eventBus events.Bus, repository Repository) Delete {
	return func(subject permission.Auditable, id ID) error {
		if err := subject.Audit(PermDelete); err != nil {
			return err
		}

		optUsr, err := repository.FindByID(id)
		if err != nil {
			return err
		}

		if err := repository.DeleteByID(id); err != nil {
			return err
		}

		// deleting is idempotent, but only an actual deletion is announced
		if optUsr.IsSome() {
			eventBus.Publish(Deleted{
				ID:        id,
				Email:     optUsr.Unwrap().Email,
				DeletedAt: time.Now(),
			})
		}

		return nil
	}
}
