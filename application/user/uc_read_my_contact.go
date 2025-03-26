// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
)

func NewReadMyContact(users Repository) ReadMyContact {
	return func(subject AuditableUser) (Contact, error) {
		if !subject.Valid() {
			return Contact{}, noLoginErr
		}

		optUser, err := users.FindByID(subject.ID())
		if err != nil {
			return Contact{}, fmt.Errorf("users.FindByID failed: %w", err)
		}

		if optUser.IsNone() {
			return Contact{}, noLoginErr
		}

		return optUser.Unwrap().Contact, nil
	}
}
