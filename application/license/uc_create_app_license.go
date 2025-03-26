// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package license

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"sync"
)

func NewCreateAppLicense(mutex *sync.Mutex, repo AppLicenseRepository) CreateAppLicense {
	return func(subject permission.Auditable, license AppLicense) (ID, error) {
		if err := subject.Audit(PermCreateAppLicense); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if license.ID == "" {
			license.ID = data.RandIdent[ID]()
		}

		optE, err := repo.FindByID(license.ID)
		if err != nil {
			return "", err
		}

		if optE.IsSome() {
			return "", std.NewLocalizedError("App-Lizenz nicht erstellbar", fmt.Sprintf("Die Lizenz mit der ID '%v' existiert bereits.", license.ID))
		}

		return license.ID, repo.Save(license)
	}
}
