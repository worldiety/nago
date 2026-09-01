// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package data

import (
	"iter"
	"testing"

	"github.com/worldiety/option"
)

type pagingTestEntity string

func (p pagingTestEntity) Identity() pagingTestEntity {
	return p
}

func pagingTestIdents(n int) iter.Seq2[pagingTestEntity, error] {
	return func(yield func(pagingTestEntity, error) bool) {
		for i := 0; i < n; i++ {
			if !yield(pagingTestEntity(Idtos(int64(i))), nil) {
				return
			}
		}
	}
}

func pagingTestFindByID(id pagingTestEntity) (option.Opt[pagingTestEntity], error) {
	return option.Some(id), nil
}

// TestPaginate_ExactMultipleOfPageSize is a regression test for a bug, where a stale/out-of-range PageIdx that
// points exactly one page beyond the last valid page (which happens whenever the total item count is an exact
// multiple of PageSize) resulted in an empty page, instead of being clamped to the last valid page.
func TestPaginate_ExactMultipleOfPageSize(t *testing.T) {
	// 5 items, page size 5 => exactly 1 page (valid index: 0).
	page, err := Paginate[pagingTestEntity, pagingTestEntity](pagingTestFindByID, pagingTestIdents(5), PaginateOptions{
		PageIdx:  4, // stale index, e.g. left over from a different, larger data set
		PageSize: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	if page.PageCount != 1 {
		t.Fatalf("expected PageCount=1, got %d", page.PageCount)
	}

	if page.PageIdx != 0 {
		t.Fatalf("expected clamped PageIdx=0, got %d", page.PageIdx)
	}

	if len(page.Items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(page.Items))
	}
}

// TestPaginate_ShrunkToExactMultipleWithoutStore covers the case where the PageIdx is exactly equal to PageCount
// (not one page beyond it), which the previous strict "<" comparison failed to detect at all.
func TestPaginate_ShrunkToExactMultipleWithoutStore(t *testing.T) {
	// 20 items, page size 5 => exactly 4 pages (valid indices: 0..3). Requesting index 4 must clamp to 3.
	page, err := Paginate[pagingTestEntity, pagingTestEntity](pagingTestFindByID, pagingTestIdents(20), PaginateOptions{
		PageIdx:  4,
		PageSize: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	if page.PageCount != 4 {
		t.Fatalf("expected PageCount=4, got %d", page.PageCount)
	}

	if page.PageIdx != 3 {
		t.Fatalf("expected clamped PageIdx=3, got %d", page.PageIdx)
	}

	if len(page.Items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(page.Items))
	}
}

// TestPaginate_NonMultiplePageSizeStillWorks ensures the ordinary, previously already-working case is unaffected.
func TestPaginate_NonMultiplePageSizeStillWorks(t *testing.T) {
	// 3 items, page size 5 => exactly 1 page (valid index: 0).
	page, err := Paginate[pagingTestEntity, pagingTestEntity](pagingTestFindByID, pagingTestIdents(3), PaginateOptions{
		PageIdx:  4,
		PageSize: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	if page.PageCount != 1 {
		t.Fatalf("expected PageCount=1, got %d", page.PageCount)
	}

	if page.PageIdx != 0 {
		t.Fatalf("expected clamped PageIdx=0, got %d", page.PageIdx)
	}

	if len(page.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(page.Items))
	}
}

// TestPaginate_ValidPageIdxIsNotTouched ensures a valid (in-range) PageIdx is never modified.
func TestPaginate_ValidPageIdxIsNotTouched(t *testing.T) {
	page, err := Paginate[pagingTestEntity, pagingTestEntity](pagingTestFindByID, pagingTestIdents(25), PaginateOptions{
		PageIdx:  2,
		PageSize: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	if page.PageCount != 5 {
		t.Fatalf("expected PageCount=5, got %d", page.PageCount)
	}

	if page.PageIdx != 2 {
		t.Fatalf("expected PageIdx=2 to remain untouched, got %d", page.PageIdx)
	}

	if len(page.Items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(page.Items))
	}
}

