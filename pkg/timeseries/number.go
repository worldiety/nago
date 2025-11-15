// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package timeseries

import (
	"math"
	"strconv"
	"time"
)

// NA declares an integer as not available.
// We use the smallest 64bit number, as a flag for NA (Not Available). There is a bunch of pro and con
// arguments and there is no clear advantage of one over the other (pointer, optionals, bitmask-tables).
// See also https://numpy.org/neps/nep-0025-missing-data-3.html for further discussions.
const NA = math.MinInt64

// NaN is used to represent NA in [F64].
var NaN = F64(math.NaN())

type UnixMilli int64

func (t UnixMilli) NA() bool {
	return I64(t).NA()
}

// String returns a debug representation in the "timezone" UTC formatted as RFC3339.
func (t UnixMilli) String() string {
	return t.StringIn(time.UTC)
}

func (t UnixMilli) StringIn(loc *time.Location) string {
	return time.UnixMilli(int64(t)).In(loc).Format(time.RFC3339)
}

type Number interface {
	~int64 | ~float64
	NA() bool
}

// I64 describes a usual signed 64bit integer type which special semantics on certain bit patterns.
// E.g. we use the smallest 64bit number, as a flag for NA (Not Available). There is a bunch of pro and con
// arguments and there is no clear advantage of one over the other (pointer, optionals, bitmask-tables).
// See also https://numpy.org/neps/nep-0025-missing-data-3.html for further discussions.
type I64 int64

// NA returns true, if the integer holds the NA bit pattern (chosen for R compatibility).
func (i I64) NA() bool {
	return i == NA
}

func (i I64) F64() F64 {
	if i.NA() {
		return F64(math.NaN())
	}

	return F64(i)
}

func (i I64) String() string {
	if i.NA() {
		return "NA"
	}

	return strconv.Itoa(int(i))
}

// F64 is a custom floating point Number.
// Keep in mind, that floats should never be used to calculate a money value from it.
// Try to stay in pre-scaled integers, to be exact. This becomes very important with very large aggregations, when
// rounding errors, even with 64bit floats, are summing up.
type F64 float64

// NA returns true, if the float represents NaN or +/- Inf.
func (f F64) NA() bool {
	return math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) || f == NA
}
