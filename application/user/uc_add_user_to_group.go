// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/pkg/std"
	"slices"
	"sync"
)

func NewAddUserToGroup(mutex *sync.Mutex, repo Repository) AddUserToGroup {
	return func(subject AuditableUser, id ID, group group.ID) error {
		if err := subject.Audit(PermUpdateOtherGroups); err != nil {
			return err
		}

		// mutex is important, otherwise we may re-create a user accidentally
		mutex.Lock()
		defer mutex.Unlock()

		optUsr, err := repo.FindByID(id)
		if err != nil {
			return fmt.Errorf("cannot find user by id: %w", err)
		}

		if optUsr.IsNone() {
			return std.NewLocalizedError("Nutzer nicht aktualisiert", "Der Nutzer ist nicht (mehr) vorhanden.")
		}

		usr := optUsr.Unwrap()
		if slices.Contains(usr.Groups, group) {
			return nil
		}

		usr.Groups = append(usr.Groups, group)
		return repo.Save(usr)
	}
}
