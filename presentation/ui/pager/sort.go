// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package pager

import (
	"iter"
	"slices"

	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
)

type SortOptions[T any] struct {
	Cache *core.State[[]T] // if not nil, this state is used as the source for the cache using Init. To reload, use [core.State.Reset].
}

// Sort loads the entire dataset into memory and applies the comparator func on it. Using the
// SortOptions the dataset may be cached within the non-nil state. Comparator may be nil and returns the identifiers
// in the original order.
func Sort[E data.Aggregate[ID], ID ~string](findByID data.ByIDFinder[E, ID], it iter.Seq2[ID, error], comparator func(a, b E) int, opts SortOptions[E]) iter.Seq2[ID, error] {
	if comparator == nil {
		return it
	}

	loadAll := func() ([]E, error) {
		ids, err := xslices.Collect2(it)
		if err != nil {
			return nil, err
		}

		tmp := make([]E, 0, len(ids))
		for _, id := range ids {
			v, err := findByID(id)
			if err != nil {
				return nil, err
			}

			if v.IsNone() {
				continue // stale ref
			}

			tmp = append(tmp, v.Unwrap())
		}

		return tmp, nil
	}

	return func(yield func(ID, error) bool) {
		var items []E
		if opts.Cache != nil {
			var fatalErr error
			opts.Cache.Init(func() []E {
				tmp, err := loadAll()
				if err != nil {
					fatalErr = err
				}

				return tmp
			})

			if fatalErr != nil {
				yield("", fatalErr)
				return
			}

			items = opts.Cache.Get() // if init was not called, the set is just re-used as-is
		} else {
			tmp, err := loadAll() // no cache, thus reload a fresh set
			if err != nil {
				yield("", err)
				return
			}

			items = tmp
		}

		// always apply the sorting, we don't know if it is a delegate and may have changed
		slices.SortFunc(items, comparator)

		for _, item := range items {
			if !yield(item.Identity(), nil) {
				return
			}
		}
	}

}
