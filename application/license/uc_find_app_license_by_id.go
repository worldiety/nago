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

func NewFindAppLicenseByID(repo AppLicenseRepository) FindAppLicenseByID {
	return func(subject permission.Auditable, id ID) (std.Option[AppLicense], error) {
		if err := subject.Audit(PermFindAppLicenseByID); err != nil {
			return std.None[AppLicense](), err
		}

		return repo.FindByID(id)
	}
}
