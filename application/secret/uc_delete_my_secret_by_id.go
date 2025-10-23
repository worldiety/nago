// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
)

func NewDeleteMySecretByID(bus events.Bus, repository Repository) DeleteMySecretByID {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermDeleteMySecretByID); err != nil {
			return err
		}

		if err := repository.DeleteByID(id); err != nil {
			return err
		}

		bus.Publish(Deleted{Secret: id})
		return nil
	}
}
