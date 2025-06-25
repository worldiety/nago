// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/auth"
	"sync"
	"time"
)

func NewUpdateEntryTransformed(mutex *sync.Mutex, repo EntryRepository) UpdateEntryTransformed {
	return func(subject auth.Subject, id Key, transformed *jsonptr.Obj) error {
		return updateEntry(mutex, repo, subject, id, PermUpdateEntryTransformed, func(entry Entry) (Entry, error) {
			entry.Transformed = transformed
			entry.LastModBy = subject.ID()
			entry.LastModAt = time.Now()
			return entry, nil
		})
	}
}
