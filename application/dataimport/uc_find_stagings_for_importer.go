// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"iter"

	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

func NewFindStagingsForImporter(repo StagingRepository) FindStagingsForImporter {
	return func(subject auth.Subject, imp importer.ID) iter.Seq2[Staging, error] {
		return func(yield func(Staging, error) bool) {
			for staging, err := range repo.All() {
				if err != nil {
					if !yield(staging, err) {
						return
					}

					continue
				}

				if staging.Importer != imp {
					continue
				}

				if !subject.HasResourcePermission(rebac.Namespace(repo.Name()), rebac.Instance(staging.ID), PermFindStaging) && (staging.CreatedBy != subject.ID() || staging.CreatedBy == "") {
					continue
				}

				if !yield(staging, nil) {
					return
				}
			}
		}
	}
}
