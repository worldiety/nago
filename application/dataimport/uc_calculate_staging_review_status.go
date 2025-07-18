// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import "go.wdy.de/nago/auth"

func NewCalculateStagingReviewStatus(repoStage StagingRepository, repo EntryRepository) CalculateStagingReviewStatus {
	return func(subject auth.Subject, staging SID, opts CalculateStagingReviewStatusOptions) (StagingReviewStatus, error) {
		if err := subject.AuditResource(repoStage.Name(), string(staging), PermCalculateStagingReviewStatus); err != nil {
			return StagingReviewStatus{}, err
		}

		var stat StagingReviewStatus
		var queryCurrent Key
		var queryPrev Key
		var queryNext Key

		var lastEntry *Entry

		for entry, err := range repo.FindAllByPrefix(Key(staging) + "/") {
			if err != nil {
				return StagingReviewStatus{}, err
			}

			if queryCurrent != "" && queryNext == "" {
				queryNext = entry.ID
			}

			if opts.Position != "" && entry.ID == opts.Position {
				queryCurrent = entry.ID
				if lastEntry != nil {
					queryPrev = lastEntry.ID
				}
			}

			lastEntry = &entry

			stat.Total++
			if entry.Confirmed || entry.Imported {
				stat.Confirmed++
			}

			if entry.Ignored {
				stat.Ignored++
			}

			if entry.Imported {
				stat.Imported++
			}

		}

		stat.CurrentEntry = opts.Position
		stat.NextEntry = queryNext
		stat.PreviousEntry = queryPrev

		return stat, nil
	}
}
