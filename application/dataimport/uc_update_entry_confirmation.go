// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"go.wdy.de/nago/auth"
	"sync"
	"time"
)

func NewUpdateEntryConfirmation(mutex *sync.Mutex, repo EntryRepository) UpdateEntryConfirmation {
	return func(subject auth.Subject, id Key, confirmed bool) error {
		return updateEntry(mutex, repo, subject, id, PermUpdateEntryConfirmation, func(entry Entry) (Entry, error) {
			entry.Confirmed = confirmed
			entry.Ignored = false // note, that not confirming something means not that it has been ignored, but ignoring and confirming at the same time are not allowed
			entry.ImportedAt = time.Time{}
			entry.ImportedError = ""
			entry.Imported = false
			return entry, nil
		})
	}
}
