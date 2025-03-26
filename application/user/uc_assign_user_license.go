// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
	"os"
	"slices"
	"sync"
)

func NewAssignUserLicense(mutex *sync.Mutex, usersRepo Repository, count CountAssignedUserLicense, findLicByID license.FindUserLicenseByID) AssignUserLicense {
	return func(auditable permission.Auditable, id ID, lic license.ID) (bool, error) {
		if err := auditable.Audit(PermAssignUserLicense); err != nil {
			return false, err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optUsr, err := usersRepo.FindByID(id)
		if err != nil {
			return false, err
		}

		if optUsr.IsNone() {
			return false, os.ErrNotExist
		}

		usr := optUsr.Unwrap()

		if slices.Contains(usr.Licenses, lic) {
			// already assigned, nothing to do
			return true, nil
		}

		// lets count and check
		used, err := count(SU(), lic)
		if err != nil {
			return false, err
		}

		optLic, err := findLicByID(SU(), lic)
		if err != nil {
			return false, err
		}

		if optLic.IsNone() {
			return false, os.ErrNotExist
		}

		if used >= optLic.Unwrap().MaxUsers {
			return false, std.NewLocalizedError("Lizenzkontingent erschöpft", fmt.Sprintf("Die maximale Anzahl von %d Lizenzen von '%s' wurde bereits zugewiesen.", optLic.Unwrap().MaxUsers, optLic.Unwrap().Name))
		}

		usr.Licenses = append(usr.Licenses, lic)

		if err := usersRepo.Save(usr); err != nil {
			return false, err
		}

		return true, nil
	}
}
