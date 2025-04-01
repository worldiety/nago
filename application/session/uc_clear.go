// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import "sync"

func NewClear(mutex *sync.Mutex, repo Repository) Clear {
	return func() error {
		mutex.Lock()
		defer mutex.Unlock()

		return repo.DeleteAll()
	}
}
