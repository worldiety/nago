// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package num provides the conversions of the gostrict subset, which are defined for every value. A conversion
// of Go like int32(f) is implementation specific for NaN and values out of range, so the subset forbids it.
//
//gostrict:header
package num

import "math"

// Trunc32 truncates f towards zero. It returns 0 for NaN and the nearest bound for values beyond the range of
// int32, e.g. for ±Inf.
func Trunc32(f float64) int32 {
	switch {
	case f != f:
		return 0
	case f >= math.MaxInt32:
		return math.MaxInt32
	case f <= math.MinInt32:
		return math.MinInt32
	}
	return int32(f)
}

// TruncU32 truncates f towards zero. It returns 0 for NaN and negative values and the maximum of uint32 for
// values beyond it.
func TruncU32(f float64) uint32 {
	switch {
	case f != f, f <= 0:
		return 0
	case f >= math.MaxUint32:
		return math.MaxUint32
	}
	return uint32(f)
}

// Round32 rounds f half away from zero like math.Round and converts it like [Trunc32].
func Round32(f float64) int32 {
	return Trunc32(math.Round(f))
}
