// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package data

import (
	"fmt"
	"iter"
	"log/slog"
	"math"
)

type PaginateOptions struct {
	// PageIdx is zero-based. Defaults to 0 as the first page.
	PageIdx int
	// PageSize defaults to 50.
	PageSize int
	// MaxResults limits the total number of entries before page size is evaluated and all id are evaluated.
	// If 0 no limit is applied.
	MaxResults int
	// IgnoreErrors silently ignores any errors but prints them into the log.
	IgnoreErrors bool
}

// Page wraps a set of loaded items.
type Page[E any] struct {
	Items []E
	// PageIdx is zero based. Note that this may be different from the requested page, if the page would be behind the dataset
	PageIdx   int
	PageSize  int
	PageCount int
	// Total is the number of all available entries.
	Total int
}

// Paginate requires the read repository which is used to resolve the items on a page.
// Technically, if no MaxResults is given, all identifiers are loaded into memory to calculate the actual paging.
// Afterward, just the items for the required page are loaded from the repository. See also [Filter] to combine
// a paging based on an ID or Aggregate Filter.
func Paginate[E Aggregate[ID], ID IDType](findByID ByIDFinder[E, ID], it iter.Seq2[ID, error], opts PaginateOptions) (Page[E], error) {
	if opts.PageSize < 1 {
		opts.PageSize = 50
	}

	if opts.PageIdx < 0 {
		opts.PageIdx = 0
	}

	var idents []ID
	for id, err := range it {
		if err != nil {
			if opts.IgnoreErrors {
				slog.Error("failed to paginate: cannot collect ids", "err", err.Error())
				continue
			}

			return Page[E]{}, fmt.Errorf("failed to paginate: cannot collect ids: %w", err)

		}

		idents = append(idents, id)
		if opts.MaxResults > 0 {
			if len(idents) >= opts.MaxResults {
				break
			}
		}
	}

	var page Page[E]
	page.PageCount = int(math.Ceil(float64(len(idents)) / float64(opts.PageSize)))
	page.PageIdx = opts.PageIdx
	page.Total = len(idents)
	page.PageSize = opts.PageSize

	if len(idents) == 0 {
		return page, nil
	}
	if len(idents) < opts.PageIdx*opts.PageSize {
		// this happens e.g. if the UI requests e.g. the second page and then applies a filter, which will cause a drop
		// of entries below the entire result set
		opts.PageIdx = len(idents) / opts.PageSize
		page.PageIdx = opts.PageIdx
	}

	offsetStart := min(opts.PageIdx*opts.PageSize, len(idents))
	offsetEnd := min(offsetStart+opts.PageSize, len(idents))
	idents = idents[offsetStart:offsetEnd]

	entries := make([]E, 0, len(idents))
	for _, ident := range idents {
		optEnt, err := findByID(ident)
		if err != nil {
			if opts.IgnoreErrors {
				slog.Error("failed to paginate: cannot find entry", "err", err.Error())
				continue
			}

			return Page[E]{}, fmt.Errorf("failed to paginate: cannot find entry: %w", err)
		}

		if optEnt.IsNone() {
			// we have no transaction scope, thus usually it is not an error, that ids are gone in the meantime.
			continue
		}

		entries = append(entries, optEnt.Unwrap())
	}

	page.Items = entries

	return page, nil
}
