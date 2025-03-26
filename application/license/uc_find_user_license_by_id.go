// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package license

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
)

func NewFindUserLicenseByID(repo UserLicenseRepository) FindUserLicenseByID {
	return func(subject permission.Auditable, id ID) (std.Option[UserLicense], error) {
		if err := subject.Audit(PermFindUserLicenseByID); err != nil {
			return std.None[UserLicense](), err
		}

		return repo.FindByID(id)
	}
}
