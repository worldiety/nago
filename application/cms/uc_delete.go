// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
	"sync"
)

func NewDelete(mutex *sync.Mutex, slugs *concurrent.RWMap[Slug, ID], repo Repository) Delete {
	return func(subject auth.Subject, id ID) error {
		if err := subject.AuditResource(repo.Name(), string(id), PermDelete); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		slugs.DeleteFunc(func(slug Slug, otherId ID) bool {
			return id == otherId
		})

		return repo.DeleteByID(id)
	}
}
