// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package license

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"sync"
)

func NewUpsertAppLicense(mutex *sync.Mutex, repo AppLicenseRepository) UpsertAppLicense {
	return func(subject permission.Auditable, license AppLicense) (ID, error) {
		if err := subject.Audit(PermCreateAppLicense); err != nil {
			return "", err
		}

		if err := subject.Audit(PermUpdateAppLicense); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if license.ID == "" {
			license.ID = data.RandIdent[ID]()
		}

		// keep the enabled flag and just update all other properties
		optLicense, err := repo.FindByID(license.ID)
		if err != nil {
			return "", err
		}

		if optLicense.IsSome() {
			license.Enabled = optLicense.Unwrap().Enabled
		}

		return license.ID, repo.Save(license)
	}
}
