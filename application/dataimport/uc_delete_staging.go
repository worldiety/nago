// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"go.wdy.de/nago/auth"
	"log/slog"
)

func NewDeleteStaging(repoStaging StagingRepository, repoEntries EntryRepository) DeleteStaging {
	return func(subject auth.Subject, staging SID) error {
		if err := subject.AuditResource(repoStaging.Name(), string(staging), PermDeleteStaging); err != nil {
			return err
		}

		if err := repoStaging.DeleteByID(staging); err != nil {
			return err
		}

		slog.Info("deleted staging", "staging", staging, "user", subject.ID())

		count := 0
		for key, err := range repoEntries.IdentifiersByPrefix(Key(string(staging) + "/")) {
			if err != nil {
				return err
			}

			if err := repoEntries.DeleteByID(key); err != nil {
				return err
			}

			count++
		}

		slog.Info("deleted staging entries", "staging", staging, "user", subject.ID(), "entries", count)

		return nil
	}
}
