// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"sync"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewAddRunConfiguration(mutex *sync.Mutex, repo Repository) AddRunConfiguration {
	return func(subject auth.Subject, pid ID, configuration RunConfiguration) error {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(pid), PermAddRunConfiguration); err != nil {
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
		configuration.ID = data.RandIdent[string]()
		prj.RunConfigurations = append(prj.RunConfigurations, configuration)
		return repo.Save(prj)
	}
}
