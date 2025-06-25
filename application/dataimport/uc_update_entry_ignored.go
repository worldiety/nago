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

func NewUpdateEntryIgnored(mutex *sync.Mutex, repo EntryRepository) UpdateEntryIgnored {
	return func(subject auth.Subject, id Key, ignored bool) error {
		return updateEntry(mutex, repo, subject, id, PermUpdateEntryIgnored, func(entry Entry) (Entry, error) {
			entry.Ignored = ignored
			entry.Confirmed = false // note, that not ignoring something means not that it has been confirmed, but ignoring and confirming at the same time are not allowed
			entry.LastModBy = subject.ID()
			entry.LastModAt = time.Now()
			return entry, nil
		})
	}
}
