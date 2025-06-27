// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package data

import "iter"

type FilterOptions[E Aggregate[ID], ID IDType] struct {
	// AcceptID can be nil, otherwise applies the predicate on each id, which is the fastest.
	AcceptID func(ID) bool

	// Accept makes a lookup for each ID and applies the predicate.
	Accept func(E) bool
}

// Filter applies the given FilterOptions to the given ID sequence and returns a filtered sequence.
// Both accept predicates are applied and concat using AND semantics. The AcceptID predicate is evaluated first.
// Referenced but not found aggregates are silently ignored.
func Filter[E Aggregate[ID], ID IDType](findByID ByIDFinder[E, ID], it iter.Seq2[ID, error], opts FilterOptions[E, ID]) iter.Seq2[ID, error] {
	return func(yield func(ID, error) bool) {
		for id, err := range it {
			if err != nil {
				if !yield(id, err) {
					return
				}

				continue
			}

			if opts.AcceptID == nil && opts.Accept == nil {
				// special case for no filter predicates is just to pass through
				if !yield(id, nil) {
					return
				}
			}

			if opts.AcceptID != nil {
				if !opts.AcceptID(id) {
					continue
				}
			}

			if opts.Accept != nil {
				optE, err := findByID(id)
				if err != nil {
					if !yield(id, err) {
						return
					}

					continue
				}

				if optE.IsNone() {
					continue
				}

				if !opts.Accept(optE.Unwrap()) {
					continue
				}

				if !yield(id, nil) {
					return
				}
			}
		}
	}
}

// FilterAndPaginate is a convenience wrapper around the Filter and Paginate functions.
func FilterAndPaginate[E Aggregate[ID], ID IDType](findByID ByIDFinder[E, ID], it iter.Seq2[ID, error], filterOpts FilterOptions[E, ID], pageOpts PaginateOptions) (Page[E], error) {
	itSeq := Filter(findByID, it, filterOpts)
	return Paginate(findByID, itSeq, pageOpts)
}
