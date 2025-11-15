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

// M4 applies the according downsampling algorithm for visualization by Uwe Jugel, Zbigniew Jerzak,
// Gregor Hackenbroich and Volker Markl. See http://www.vldb.org/pvldb/vol7/p797-jugel.pdf. It expects
// that the given db.TimeSeries is already sorted.
//
// The width determines how many buckets are created in the given time interval as defined by the time series.
// Each bucket may have a variable amount of entries, which are sampled to at most 4 values:
// the highest/lowest values and the max/min values. If these points overlap, they are only returned once, so
// at worst only one value per bucket is returned.
//
// If the width is larger than the amount of available seconds, the original points are returned.
func M4[Y Number](it iter.Seq[Point[UnixMilli, Y]], interval Range, width int) iter.Seq[Point[UnixMilli, Y]] {
	tMin, tMax, err := interval.Interval()
	if err != nil {
		panic(fmt.Errorf("invalid interval: %w", err))
	}

	timeInterval := (tMax - tMin) / UnixMilli(width)
	if timeInterval <= 0 {
		return it // return unmodified data set, because our time interval is too small
	}

	return func(yield func(Point[UnixMilli, Y]) bool) {
		left, right := Point[UnixMilli, Y]{X: NA}, Point[UnixMilli, Y]{X: NA}
		min, max := Point[UnixMilli, Y]{X: NA}, Point[UnixMilli, Y]{X: NA}
		t := tMin
		for point := range it {
			next := m4Points(point, timeInterval, &left, &right, &min, &max)
			if !next {
				set := m4UniquePoints(left, right, min, max)

				left, right = Point[UnixMilli, Y]{X: NA}, Point[UnixMilli, Y]{X: NA}
				min, max = Point[UnixMilli, Y]{X: NA}, Point[UnixMilli, Y]{X: NA}
				for _, p := range set {
					if p.X != NA {
						if !yield(p) {
							return
						}
					}
				}
			} else {
				t += timeInterval
			}
		}
	}

}

// m4Points returns the points for left, right, min and max values of the given series.
// Having secondsInWindow <= 0 is undefined.
func m4Points[Y Number](p Point[UnixMilli, Y], secondsInWindow UnixMilli, left, right, min, max *Point[UnixMilli, Y]) (next bool) {
	if left.X == NA {
		*left = p
	}

	*right = p
	if min.X == NA || p.Y < min.Y {
		*min = p
	}

	if max.X == NA || p.Y > max.Y {
		*max = p
	}

	if right.X-left.X >= secondsInWindow {
		return false
	}

	return true
}

// m4UniquePoints returns an array with 4 indices, if they are not unique. Invalid indices have the value -1.
func m4UniquePoints[Y Number](left, right, min, max Point[UnixMilli, Y]) [4]Point[UnixMilli, Y] {
	var set [4]Point[UnixMilli, Y]
	set[0] = left
	if min.X < max.X {
		// min is older than max
		set[1] = min
		set[2] = max
	} else {
		// max is older than min
		set[1] = max
		set[2] = min
	}
	set[3] = right

	// search for duplicates and set those to NA. We can short circuit because we are always monotonic (== equal
	// timestamps) or even strict monotonic (== all different timestamps).
	for i1, p1 := range set {
		for i2 := i1 + 1; i2 < len(set); i2++ {
			if p1.X == set[i2].X {
				set[i2].X = NA
			}
		}
	}

	return set
}
