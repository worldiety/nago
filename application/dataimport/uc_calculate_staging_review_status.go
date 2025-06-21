// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import "go.wdy.de/nago/auth"

func NewCalculateStagingReviewStatus(repoStage StagingRepository, repo EntryRepository) CalculateStagingReviewStatus {
	return func(subject auth.Subject, staging SID) (StagingReviewStatus, error) {
		if err := subject.AuditResource(repoStage.Name(), string(staging), PermCalculateStagingReviewStatus); err != nil {
			return StagingReviewStatus{}, err
		}

		var stat StagingReviewStatus
		for entry, err := range repo.FindAllByPrefix(Key(staging) + "/") {
			if err != nil {
				return StagingReviewStatus{}, err
			}

			stat.Total++
			if entry.Confirmed {
				stat.Confirmed++
			}

			if entry.Ignored {
				stat.Ignored++
			}

			if entry.Imported {
				stat.Imported++
			}

		}

		return stat, nil
	}
}
