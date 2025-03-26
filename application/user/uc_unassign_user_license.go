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
	"os"
	"slices"
	"sync"
)

func NewUnassignUserLicense(mutex *sync.Mutex, usersRepo Repository) UnassignUserLicense {
	return func(auditable permission.Auditable, id ID, lic license.ID) error {
		if err := auditable.Audit(PermAssignUserLicense); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optUsr, err := usersRepo.FindByID(id)
		if err != nil {
			return err
		}

		if optUsr.IsNone() {
			return os.ErrNotExist
		}

		usr := optUsr.Unwrap()

		if !slices.Contains(usr.Licenses, lic) {
			// already assigned, nothing to do
			return nil
		}

		usr.Licenses = slices.DeleteFunc(usr.Licenses, func(id license.ID) bool {
			return id == lic
		})

		return usersRepo.Save(usr)
	}
}
