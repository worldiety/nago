// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"go.wdy.de/nago/auth"
	"sync"
)

func NewDeleteByID(mutex *sync.Mutex, repo Repository) DeleteByID {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermDeleteByID); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		return repo.DeleteByID(id)
	}
}
