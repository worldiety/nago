// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"iter"

	"go.wdy.de/nago/pkg/timeseries"
)

// The read API has two surfaces per value type:
//
//   - Scan*  : the columnar batch reader. It yields decoded blocks as parallel
//     slices with zero per-point closure cost. This is the path for billion-
//     element scans; the slices are only valid for the duration of the callback
//     (they are reused between calls) — clone to retain.
//   - Iter*  : a convenience iter.Seq of timeseries.Point built on top of Scan*,
//     which composes directly with timeseries.M4 and timeseries.Series.

// ---- numeric: raw scaled int64 ----

// ScanI64 invokes fn for each decoded block of raw scaled int64 values with
// min <= ts <= max, in ascending order. ts and vals are parallel and reused
// across calls. Returning false stops iteration. Valid only for SchemeDecimal.
func (c *Column) ScanI64(min, max int64, fn func(ts []int64, vals []int64) bool) error {
	if c.schema.Scheme != SchemeDecimal {
		return errSchemeMismatch
	}
	return c.mergedRange(min, max, func(ts []int64, vals []int64, _ []string) bool {
		return fn(ts, vals)
	})
}

// ScanF64 is like ScanI64 but unscales values to float64 using the column's
// decimals. The vals slice is engine-owned scratch, valid only during fn.
func (c *Column) ScanF64(min, max int64, fn func(ts []int64, vals []float64) bool) error {
	if c.schema.Scheme != SchemeDecimal {
		return errSchemeMismatch
	}
	dec := c.schema.Decimals
	var fbuf []float64
	return c.mergedRange(min, max, func(ts []int64, vals []int64, _ []string) bool {
		if cap(fbuf) < len(vals) {
			fbuf = make([]float64, len(vals))
		}
		fbuf = fbuf[:len(vals)]
		for i, v := range vals {
			fbuf[i] = unscale(v, dec)
		}
		return fn(ts, fbuf)
	})
}

// ScanString invokes fn for each decoded block of string values. Valid for
// SchemeString and SchemeEnum (enum ids are resolved to strings).
func (c *Column) ScanString(min, max int64, fn func(ts []int64, vals []string) bool) error {
	switch c.schema.Scheme {
	case SchemeString:
		return c.mergedRange(min, max, func(ts []int64, _ []int64, strs []string) bool {
			return fn(ts, strs)
		})
	case SchemeEnum:
		var sbuf []string
		return c.mergedRange(min, max, func(ts []int64, ids []int64, _ []string) bool {
			if cap(sbuf) < len(ids) {
				sbuf = make([]string, len(ids))
			}
			sbuf = sbuf[:len(ids)]
			for i, id := range ids {
				s, _ := c.dict.lookup(uint32(id))
				sbuf[i] = s
			}
			return fn(ts, sbuf)
		})
	default:
		return errSchemeMismatch
	}
}

// ---- iterators (compose with timeseries.M4 / Series) ----

// IterI64 yields raw scaled int64 points. Composes with timeseries.M4.
func (c *Column) IterI64(min, max int64) iter.Seq[timeseries.Point[timeseries.UnixMilli, timeseries.I64]] {
	return func(yield func(timeseries.Point[timeseries.UnixMilli, timeseries.I64]) bool) {
		_ = c.ScanI64(min, max, func(ts []int64, vals []int64) bool {
			for i := range ts {
				if !yield(timeseries.Point[timeseries.UnixMilli, timeseries.I64]{
					X: timeseries.UnixMilli(ts[i]),
					Y: timeseries.I64(vals[i]),
				}) {
					return false
				}
			}
			return true
		})
	}
}

// IterF64 yields unscaled float64 points. Composes with timeseries.M4.
func (c *Column) IterF64(min, max int64) iter.Seq[timeseries.Point[timeseries.UnixMilli, timeseries.F64]] {
	return func(yield func(timeseries.Point[timeseries.UnixMilli, timeseries.F64]) bool) {
		_ = c.ScanF64(min, max, func(ts []int64, vals []float64) bool {
			for i := range ts {
				if !yield(timeseries.Point[timeseries.UnixMilli, timeseries.F64]{
					X: timeseries.UnixMilli(ts[i]),
					Y: timeseries.F64(vals[i]),
				}) {
					return false
				}
			}
			return true
		})
	}
}

// StringPoint is one (timestamp, string) pair.
type StringPoint struct {
	X timeseries.UnixMilli
	V string
}

// IterString yields string points (string or enum scheme).
func (c *Column) IterString(min, max int64) iter.Seq[StringPoint] {
	return func(yield func(StringPoint) bool) {
		_ = c.ScanString(min, max, func(ts []int64, vals []string) bool {
			for i := range ts {
				if !yield(StringPoint{X: timeseries.UnixMilli(ts[i]), V: vals[i]}) {
					return false
				}
			}
			return true
		})
	}
}

// ---- typed writers ----

// PutI64 writes/overwrites a raw scaled int64 value at ts (SchemeDecimal).
func (c *Column) PutI64(ts int64, v int64) error {
	if c.schema.Scheme != SchemeDecimal {
		return errSchemeMismatch
	}
	return c.putNumeric(ts, v)
}

// PutF64 writes/overwrites a float64 value at ts with transparent scaling to
// int64 using the column's decimals (SchemeDecimal).
func (c *Column) PutF64(ts int64, v float64) error {
	if c.schema.Scheme != SchemeDecimal {
		return errSchemeMismatch
	}
	return c.putNumeric(ts, scale(v, c.schema.Decimals))
}

// PutString writes/overwrites a string value at ts (SchemeString or SchemeEnum).
func (c *Column) PutString(ts int64, v string) error {
	if c.schema.Scheme != SchemeString && c.schema.Scheme != SchemeEnum {
		return errSchemeMismatch
	}
	return c.putString(ts, v)
}
