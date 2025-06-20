// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xslices"
	"math"
)

func NewFilterEntries(repoEntries EntryRepository) FilterEntries {
	return func(subject auth.Subject, stage SID, opts FilterEntriesOptions) (FilterEntriesPage, error) {
		if err := subject.Audit(PermFilterEntries); err != nil {
			return FilterEntriesPage{}, err
		}

		if opts.PageSize <= 1 {
			opts.PageSize = 50
		}

		if opts.Page < 0 {
			opts.Page = 0
		}

		idents, err := xslices.Collect2(repoEntries.IdentifiersByPrefix(Key(string(stage) + "/")))
		if err != nil {
			return FilterEntriesPage{}, err
		}

		var page FilterEntriesPage
		page.PageCount = int(math.Ceil(float64(len(idents)) / float64(opts.PageSize)))
		page.Page = opts.Page
		page.Count = int64(len(idents))
		page.PageSize = opts.PageSize

		if len(idents) == 0 {
			return page, nil
		}

		offsetStart := min(opts.Page*opts.PageSize, len(idents)-1)
		offsetEnd := min(offsetStart+opts.PageSize, len(idents)-1)
		idents = idents[offsetStart : offsetEnd+1]

		entries := make([]Entry, 0, len(idents))
		for _, ident := range idents {
			optEnt, err := repoEntries.FindByID(ident)
			if err != nil {
				return FilterEntriesPage{}, err
			}

			if optEnt.IsNone() {
				continue
			}

			entries = append(entries, optEnt.Unwrap())
			if opts.MaxResults > 0 && len(entries) >= opts.MaxResults {
				break
			}
		}

		page.Entries = entries

		return page, nil
	}
}
