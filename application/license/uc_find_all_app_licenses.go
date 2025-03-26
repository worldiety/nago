// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package license

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewFindAllAppLicenses(repo AppLicenseRepository) FindAllAppLicenses {
	return func(subject permission.Auditable) iter.Seq2[AppLicense, error] {
		if err := subject.Audit(PermFindAllAppLicenses); err != nil {
			return xiter.WithError[AppLicense](err)
		}

		return repo.All()
	}
}
