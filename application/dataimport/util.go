// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"os"
	"sync"
)

func updateEntry(mutex *sync.Mutex, repo EntryRepository, subject auth.Subject, id Key, perm permission.ID, fn func(entry Entry) (Entry, error)) error {
	if err := subject.AuditResource(repo.Name(), string(id), perm); err != nil {
		return err
	}

	mutex.Lock()
	defer mutex.Unlock()

	optEntry, err := repo.FindByID(id)
	if err != nil {
		return err
	}

	if optEntry.IsNone() {
		return fmt.Errorf("cannot find entry by id %s: %w", id, os.ErrNotExist)
	}

	entry, err := fn(optEntry.Unwrap())
	if err != nil {
		return err
	}

	return repo.Save(entry)
}
