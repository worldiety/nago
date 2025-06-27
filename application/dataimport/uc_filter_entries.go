// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewFilterEntries(repoEntries EntryRepository) FilterEntries {
	return func(subject auth.Subject, stage SID, opts data.PaginateOptions) (data.Page[Entry], error) {
		if err := subject.Audit(PermFilterEntries); err != nil {
			return data.Page[Entry]{}, err
		}

		return data.Paginate(repoEntries.FindByID, repoEntries.IdentifiersByPrefix(Key(string(stage)+"/")), opts)
	}
}
