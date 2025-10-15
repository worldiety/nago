// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package group

import (
	"fmt"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"

	"strings"
	"sync"
)

func NewUpsert(mutex *sync.Mutex, bus events.Bus, repo Repository) Upsert {
	return func(subject permission.Auditable, group Group) (ID, error) {
		if err := subject.Audit(PermCreate); err != nil {
			return "", err
		}

		if err := subject.Audit(PermUpdate); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		createNew := false
		if strings.TrimSpace(string(group.ID)) == "" {
			group.ID = data.RandIdent[ID]()
			createNew = true
		}

		optGroup, err := repo.FindByID(group.ID)
		if err != nil {
			return "", fmt.Errorf("cannot find group by id: %w", err)
		}

		if optGroup.IsSome() && createNew {
			return "", fmt.Errorf("random id collision on upsert creation")
		}

		if err := repo.Save(group); err != nil {
			return "", fmt.Errorf("cannot save group: %w", err)
		}

		if optGroup.IsSome() {
			bus.Publish(Updated{Group: group.ID})
		}

		return group.ID, nil
	}
}
