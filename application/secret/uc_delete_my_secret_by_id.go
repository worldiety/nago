// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"fmt"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
)

func NewDeleteMySecretByID(bus events.Bus, repository Repository) DeleteMySecretByID {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermDeleteMySecretByID); err != nil {
			return err
		}

		optSecret, err := repository.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot find secret: %w", err)
		}

		if optSecret.IsNone() {
			// nothing to do, deleting is idempotent
			return nil
		}

		if !optSecret.Unwrap().HasAccess(subject) {
			return AccessDeniedErr
		}

		if err := repository.DeleteByID(id); err != nil {
			return err
		}

		bus.Publish(Deleted{Secret: id})
		return nil
	}
}
