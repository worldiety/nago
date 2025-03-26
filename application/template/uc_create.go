// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package template

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"sync"
)

func NewCreate(mutex *sync.Mutex, repository Repository) Create {
	return func(subject auth.Subject, project Project) (ID, error) {
		if err := subject.Audit(PermCreate); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if project.ID == "" {
			project.ID = data.RandIdent[ID]()
		}

		optPrj, err := repository.FindByID(project.ID)
		if err != nil {
			return "", fmt.Errorf("cannot find project by id: %s", project.ID)
		}

		if optPrj.IsSome() {
			return "", fmt.Errorf("project already exists: %s", project.ID)
		}

		return project.ID, repository.Save(project)
	}
}
