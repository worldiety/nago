// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package timeseries

import (
	"reflect"
	"sort"
	"testing"
)

func TestTimeSeries_Validate(t *testing.T) {
	ts := Series[I64]{
		{1, 2},
		{2, 3},
		{0, 4},
	}

	if err := ts.Validate(); err == nil {
		t.Fatal("must fail")
	}

	sort.Sort(ts)

	if err := ts.Validate(); err != nil {
		t.Fatal(err)
	}

	ts = Series[I64]{
		{3, 4},
		{3, 2},
		{2, 3},
	}

	sort.Sort(ts)

	if err := ts.Validate(); err == nil {
		t.Fatal("expected double key failure")
	}

	ts.Fuse()

	if err := ts.Validate(); err != nil {
		t.Fatal(err)
	}

	if len(ts) != 2 {
		t.Fatal("expected length of 2")
	}

	if ts[1].X != 3 || ts[1].Y != 2 {
		t.Fatalf("last value is wrong: %v", ts[1])
	}

	ts = Series[I64]{
		{3, 3},
		{3, 2},
		{3, 1},
	}
	ts.Fuse()

	if len(ts) != 1 {
		t.Fatal(ts)
	}

	if ts[0].X != 3 || ts[0].Y != 1 {
		t.Fatalf("last value is wrong: %v", ts[0])
	}

	ts = Series[I64]{
		{3, 1},
		{3, 2},
		{3, 3},
		{2, 4},
		{2, 5},
		{2, 6},
		{1, 7},
		{1, 8},
		{1, 9},
	}
	ts.Fuse()

	if !reflect.DeepEqual(ts, Series[I64]{
		{1, 9},
		{2, 6},
		{3, 3},
	}) {
		t.Fatal(ts)
	}

	ts = Series[I64]{
		{1, 1},
		{2, 2},
		{3, 3},
		{3, 4},
		{3, 5},
		{3, 6},
		{3, 7},
		{3, 8},
		{4, 9},
	}
	ts.Fuse()

	if !reflect.DeepEqual(ts, Series[I64]{
		{1, 1},
		{2, 2},
		{3, 8},
		{4, 9},
	}) {
		t.Fatal(ts)
	}
}

func TestTimeSeries_Delete(t *testing.T) {
	// simple one
	ts := Series[I64]{
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
	}
	ts.Delete(3, 3)

	if !reflect.DeepEqual(ts, Series[I64]{
		{1, 1},
		{2, 2},
		{4, 4},
	}) {
		t.Fatal(ts)
	}

	// simple range
	ts = Series[I64]{
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
	}
	ts.Delete(3, 4)

	if !reflect.DeepEqual(ts, Series[I64]{
		{1, 1},
		{2, 2},
		{5, 5},
	}) {
		t.Fatal(ts)
	}

	// check idempotent delete
	ts.Delete(3, 4)

	if !reflect.DeepEqual(ts, Series[I64]{
		{1, 1},
		{2, 2},
		{5, 5},
	}) {
		t.Fatal(ts)
	}

	// out of range: before
	ts = Series[I64]{
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
	}
	ts.Delete(0, 1)

	if !reflect.DeepEqual(ts, Series[I64]{
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
	}) {
		t.Fatal(ts)
	}

	// out of range: after
	ts = Series[I64]{
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
	}
	ts.Delete(6, 7)

	if !reflect.DeepEqual(ts, Series[I64]{
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
	}) {
		t.Fatal(ts)
	}

	// delete first
	ts = Series[I64]{
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
	}
	ts.Delete(0, 2)

	if !reflect.DeepEqual(ts, Series[I64]{
		{3, 3},
		{4, 4},
		{5, 5},
	}) {
		t.Fatal(ts)
	}

	// delete last
	ts = Series[I64]{
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
	}
	ts.Delete(5, 7)

	if !reflect.DeepEqual(ts, Series[I64]{
		{2, 2},
		{3, 3},
		{4, 4},
	}) {
		t.Fatal(ts)
	}

	// delete empty
	ts = Series[I64]{}
	ts.Delete(5, 7)

	if !reflect.DeepEqual(ts, Series[I64]{}) {
		t.Fatal(ts)
	}

}
