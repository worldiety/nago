// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"go.wdy.de/nago/auth"
)

func NewDeleteMySecretByID(repository Repository) DeleteMySecretByID {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermDeleteMySecretByID); err != nil {
			return err
		}

		return repository.DeleteByID(id)
	}
}
