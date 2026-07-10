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
)

// M4 applies a visualization-oriented downsampling. It expects the input iterator
// to be sorted ascending by time.
//
// The time span [tMin,tMax] defined by interval is divided into width buckets of
// equal width Δ = (tMax-tMin)/width. A point at time t falls into the fixed grid
// bucket floor((t-tMin)/Δ) (t == tMax is clamped into the last bucket). For each
// non-empty bucket M4 emits at most four points — the ones with the smallest
// time (first), the largest time (last), the minimum value and the maximum value
// — deduplicated and in ascending time order. These four per pixel column are
// exactly what is needed to render a line chart without visual loss.
//
// This implementation is a single streaming pass with O(1) memory: because the
// input is time-sorted, only the four accumulators of the currently open bucket
// are ever held, and a bucket is flushed as soon as a point crosses into a later
// bucket. It therefore scales to billions of points at constant memory.
//
// If Δ <= 0 (width larger than the available time span) the input is returned
// unchanged.
func M4[Y Number](it iter.Seq[Point[UnixMilli, Y]], interval Range, width int) iter.Seq[Point[UnixMilli, Y]] {
	tMin, tMax, err := interval.Interval()
	if err != nil {
		panic(fmt.Errorf("invalid interval: %w", err))
	}

	if width <= 0 {
		return it
	}
	delta := (tMax - tMin) / UnixMilli(width)
	if delta <= 0 {
		return it // time span too small to bucket; return unmodified data set
	}
	lastBucket := int64(width) - 1

	return func(yield func(Point[UnixMilli, Y]) bool) {
		var (
			curBucket int64 = -1
			have      bool
			first     Point[UnixMilli, Y]
			last      Point[UnixMilli, Y]
			min       Point[UnixMilli, Y]
			max       Point[UnixMilli, Y]
		)

		// flushBucket emits the current bucket's first, min, max and last points
		// in ascending time order, dropping duplicates. Returns false if the
		// consumer stopped.
		flushBucket := func() bool {
			if !have {
				return true
			}
			return emitM4Bucket(first, last, min, max, yield)
		}

		for p := range it {
			// assign the fixed grid bucket; clamp out-of-range into [0,lastBucket]
			b := int64((p.X - tMin) / delta)
			if b < 0 {
				b = 0
			}
			if b > lastBucket {
				b = lastBucket
			}

			if !have || b != curBucket {
				if !flushBucket() {
					return
				}
				curBucket = b
				have = true
				first, last, min, max = p, p, p, p
				continue
			}

			last = p
			if p.Y < min.Y {
				min = p
			}
			if p.Y > max.Y {
				max = p
			}
		}

		flushBucket()
	}
}

// emitM4Bucket yields the first, min, max and last points of a bucket in
// ascending time order without duplicates. At most four, at least one, point is
// produced. Returns false if the consumer stopped.
func emitM4Bucket[Y Number](first, last, min, max Point[UnixMilli, Y], yield func(Point[UnixMilli, Y]) bool) bool {
	// order the four candidates by time: first is the earliest and last the
	// latest by construction; min and max fall in between in either order.
	var ordered [4]Point[UnixMilli, Y]
	ordered[0] = first
	if min.X <= max.X {
		ordered[1] = min
		ordered[2] = max
	} else {
		ordered[1] = max
		ordered[2] = min
	}
	ordered[3] = last

	prevX := UnixMilli(NA)
	for i, p := range ordered {
		// skip duplicates by timestamp; the array is non-decreasing in time.
		if i > 0 && p.X == prevX {
			continue
		}
		prevX = p.X
		if !yield(p) {
			return false
		}
	}
	return true
}
