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
	"strings"
	"sync"
)

func NewUpsert(mutex *sync.Mutex, repo Repository) Upsert {
	return func(subject permission.Auditable, role Role) (ID, error) {
		if err := subject.Audit(PermCreate); err != nil {
			return "", err
		}

		if err := subject.Audit(PermUpdate); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		createNew := false
		if strings.TrimSpace(string(role.ID)) == "" {
			role.ID = data.RandIdent[ID]()
			createNew = true
		}

		optGroup, err := repo.FindByID(role.ID)
		if err != nil {
			return "", fmt.Errorf("cannot find group by id: %w", err)
		}

		if optGroup.IsSome() && createNew {
			return "", fmt.Errorf("random id collision on upsert creation")
		}

		return role.ID, repo.Save(role)
	}
}
