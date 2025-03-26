// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package role

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"strings"
	"sync"
)

func NewCreate(mutex *sync.Mutex, repo Repository) Create {
	return func(subject permission.Auditable, role Role) (ID, error) {
		if err := subject.Audit(PermCreate); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if strings.TrimSpace(string(role.ID)) == "" {
			role.ID = data.RandIdent[ID]()
		}

		optGroup, err := repo.FindByID(role.ID)
		if err != nil {
			return "", fmt.Errorf("cannot find group by id: %w", err)
		}

		if optGroup.IsSome() {
			return "", std.NewLocalizedError("Ungültige EID", "Eine Gruppe mit derselben EID ist bereits vorhanden.")
		}

		return role.ID, repo.Save(role)
	}
}
