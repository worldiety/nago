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
	"go.wdy.de/nago/pkg/std"
	"sync"
)

func NewUpdateAppLicense(mutex *sync.Mutex, repo AppLicenseRepository) UpdateAppLicense {
	return func(subject permission.Auditable, license AppLicense) error {
		if err := subject.Audit(PermUpdateAppLicense); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optE, err := repo.FindByID(license.ID)
		if err != nil {
			return err
		}

		if optE.IsNone() {
			return std.NewLocalizedError("App-Lizenz nicht aktualisierbar", fmt.Sprintf("Die Lizenz mit der ID '%v' existiert nicht.", license.ID))
		}

		return repo.Save(license)
	}
}
