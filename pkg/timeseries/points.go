// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package timeseries

import (
	"fmt"
	"iter"
	"sort"
)

// PI is an alias for Time series Point
type PI = Point[UnixMilli, I64]
type PF = Point[UnixMilli, F64]

// Points as an arbitrary slice of Point values.
type Points[T, D Number] []Point[T, D]

// A Series is a slice of Points whose X type is a UTC type and whose Y value is optional.
type Series[D Number] Points[UnixMilli, D]

// An ISeries is just an alias for a time Series of I64 values.
type ISeries = Series[I64]

func (t Series[D]) Len() int {
	return len(t)
}

func (t Series[D]) Less(i, j int) bool {
	return t[i].X < t[j].X
}

func (t Series[D]) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Validate checks that the Series is valid, which means that it is sorted ascending and does not contain
// duplicates.
func (t Series[D]) Validate() error {
	if len(t) < 2 {
		return nil
	}

	last := t[0].X
	for i := 1; i < len(t); i++ {
		cur := t[i].X
		if cur <= last {
			return fmt.Errorf("invalid Series: expected t[%d] > t[%d] but %d <= %d", i, i-1, cur, last)
		}

		last = cur
	}

	return nil
}

// Delete will purge all values which are (>= min and <= max). The length is truncated but capacity is unchanged.
// It expects a sorted time series, exactly as produced by Fuse. The result on unsorted or duplicated data is undefined.
// It locates the positions of min and max and performs a single memcopy to fill the gap.
//
//goland:noinspection GoMixedReceiverTypes
func (t *Series[D]) Delete(min, max UnixMilli) {
	if t == nil || len(*t) == 0 {
		return
	}

	if min > max {
		min, max = max, min
	}

	idxLeft := sort.Search(len(*t), func(i int) bool {
		return (*t)[i].X >= min
	})

	// min is larger than largest timestamp
	if idxLeft >= len(*t) {
		return
	}

	idxRight := sort.Search(len(*t), func(i int) bool {
		return (*t)[i].X >= max
	})

	// it is fine that the upper bound exceeds the max of our shards, just clamp it
	if idxRight >= len(*t) {
		idxRight = len(*t) - 1
	}

	// idxRight might be offset, because search only finds the insertion point
	for (*t)[idxRight].X > max && idxRight > 0 {
		idxRight--
	}

	// max is smaller than the smallest time, nothing to delete
	if idxRight == 0 && (*t)[0].X > max {
		return
	}

	// special case: no elements in range: lower bound and upper bound point to the same index (to insert)
	if idxLeft == idxRight && (*t)[idxRight].X > max {
		return
	}

	// copy everything from right to left, using the distance and truncate slice by that
	copy((*t)[idxLeft:], (*t)[idxRight+1:])
	*t = (*t)[:len(*t)-(idxRight-idxLeft)-1]
}

// Fuse will purge duplicate entries. It first ensures correct sorting and removes each duplicate time value, until
// the most recent one. It may change the length of the Series. The implementation is optimized, so that it
// performs an in-place substitution without the usual "memmove" slice tricks, which will otherwise result
// in enormous runtime penalties.
//
//goland:noinspection GoMixedReceiverTypes
func (t *Series[D]) Fuse() {
	if len(*t) < 2 {
		return
	}

	if err := t.Validate(); err != nil {
		sort.Stable(t)

		// nothing to do, sorting fixed it
		if err = t.Validate(); err == nil {
			return
		}
	}

	// the following code is at least 130x faster in our real-world conditions
	// because it has just a linear effort, instead n*(pos-length) bytes for n duplicates,
	// e.g. this takes now 230-290ms instead of 30s+.
	a := *t
	j := 1 // sliding window for look-ahead
	for i := 1; i < len(a); i++ {
		if a[i].X != a[i-1].X {
			a[j] = a[i]
			j++
		} else {
			// flexible window for look-back: push value to the beginning
			for k := j - 1; k >= 0; k-- {
				if a[k].X == a[i].X {
					a[k].Y = a[i].Y
				} else {
					break
				}
			}
		}
	}

	a = a[:j]

	*t = a
}

// First returns the first Point.
func (t Series[D]) First() Point[UnixMilli, D] {
	if len(t) == 0 {
		return Point[UnixMilli, D]{
			Y: NA,
		}
	}

	return t[0]
}

// Last returns the last Point.
func (t Series[D]) Last() Point[UnixMilli, D] {
	if len(t) == 0 {
		return Point[UnixMilli, D]{
			Y: NA,
		}
	}

	return t[len(t)-1]
}

func (t Series[D]) All() iter.Seq[Point[UnixMilli, D]] {
	return func(yield func(Point[UnixMilli, D]) bool) {
		for _, p := range t {
			if !yield(p) {
				return
			}
		}
	}
}
