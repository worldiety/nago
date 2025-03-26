// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
)

func NewCountAssignedUserLicense(users Repository) CountAssignedUserLicense {
	return func(auditable permission.Auditable, id license.ID) (int, error) {
		if err := auditable.Audit(PermCountAssignedUserLicense); err != nil {
			return 0, err
		}

		// deadlock note: keep mutex out of scope, read-write use cases ensure the global mutex
		var count int
		for user, err := range users.All() {
			if err != nil {
				return 0, err
			}

			for _, lic := range user.Licenses {
				if lic == id {
					count++
					break
				}
			}
		}

		return count, nil
	}
}
