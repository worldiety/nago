// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package agent

import (
	"fmt"
	"os"
	"sync"

	"go.wdy.de/nago/auth"
)

func NewUpdate(mutex *sync.Mutex, repo Repository) Update {
	return func(subject auth.Subject, ag Agent) error {
		if err := subject.AuditResource(repo.Name(), string(ag.ID), PermUpdate); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optAg, err := repo.FindByID(ag.ID)
		if err != nil {
			return err
		}

		if optAg.IsNone() {
			return fmt.Errorf("agent id %q not found: %w", ag.ID, os.ErrNotExist)
		}

		return repo.Save(ag)
	}
}
