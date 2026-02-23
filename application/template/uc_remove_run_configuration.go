// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"slices"
	"sync"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

func NewRemoveRunConfiguration(mutex *sync.Mutex, repo Repository) RemoveRunConfiguration {
	return func(subject auth.Subject, pid ID, nameOrId string) error {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(pid), PermRemoveRunConfiguration); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optPrj, err := repo.FindByID(pid)
		if err != nil {
			return err
		}

		if optPrj.IsNone() {
			return nil
		}

		prj := optPrj.Unwrap()
		prj.RunConfigurations = slices.DeleteFunc(prj.RunConfigurations, func(e RunConfiguration) bool {
			return e.ID == nameOrId || e.Name == nameOrId
		})

		return repo.Save(prj)
	}
}
