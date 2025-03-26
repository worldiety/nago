// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"go.wdy.de/nago/auth"
	"os"
	"sync"
)

func NewUpdate(mutex *sync.Mutex, repo Repository) Update {
	return func(subject auth.Subject, circle Circle) error {
		if err := subject.Audit(PermUpdate); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optCircle, err := repo.FindByID(circle.ID)
		if err != nil {
			return err
		}

		if optCircle.IsNone() {
			return os.ErrNotExist
		}

		return repo.Save(circle)
	}
}
