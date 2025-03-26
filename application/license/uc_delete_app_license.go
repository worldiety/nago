// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package license

import (
	"go.wdy.de/nago/application/permission"
	"sync"
)

func NewDeleteAppLicense(mutex *sync.Mutex, repo AppLicenseRepository) DeleteAppLicense {
	return func(subject permission.Auditable, id ID) error {
		if err := subject.Audit(PermDeleteAppLicense); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		return repo.DeleteByID(id)
	}
}
