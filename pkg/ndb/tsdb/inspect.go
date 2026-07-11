// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

// ColumnStats is cheap, metadata-only information about one column. All fields
// are derived from the in-memory chunk set and the recovered max timestamp
// without scanning any block or decoding any point, so it is safe to call over
// columns holding billions of points.
//
// A point count is deliberately not provided: tsdb does not track it and it
// would require a full scan. Use a bounded Scan* if an exact count is needed.
type ColumnStats struct {
	// Scheme and Decimals are the column's immutable schema.
	Scheme   Scheme
	Decimals uint8
	// Chunks is the number of sealed chunk files (the open pending chunk is not
	// counted here; its data is still reflected in MaxMillis).
	Chunks int
	// Bytes is the total on-disk size of the sealed chunks (finalized files).
	Bytes int64
	// MinMillis / MaxMillis bound the time range that currently holds data
	// (inclusive, unix milliseconds). Valid only when HasData is true.
	MinMillis int64
	MaxMillis int64
	// HasData reports whether the column holds any point (sealed or buffered).
	HasData bool
}

// Stats returns cheap, metadata-only statistics for the column. It reads only
// the in-memory chunk set and the recovered max timestamp under the column
// lock; it never decodes a block, so it is O(chunks), not O(points).
func (c *Column) Stats() ColumnStats {
	c.mu.Lock()
	defer c.mu.Unlock()

	st := ColumnStats{
		Scheme:   c.schema.Scheme,
		Decimals: c.schema.Decimals,
		Chunks:   len(c.chunks),
	}

	var haveMin bool
	setMin := func(v int64) {
		if !haveMin || v < st.MinMillis {
			st.MinMillis = v
			haveMin = true
		}
	}
	setMax := func(v int64) {
		if !st.HasData || v > st.MaxMillis {
			st.MaxMillis = v
		}
		st.HasData = true
	}

	for _, ci := range c.chunks {
		st.Bytes += ci.sizeBytes
		setMin(ci.minMillis)
		setMax(ci.maxMillis)
	}

	// The append buffer / head may hold data beyond the sealed chunks.
	if hmin, ok := c.head.minTS(); ok {
		setMin(hmin)
	}
	if c.haveCurMax {
		setMin(c.curMax) // ensures min is set even if only buffered data exists
		setMax(c.curMax)
	}

	if !haveMin {
		st.HasData = false
	}
	return st
}

// Count returns the exact number of points currently stored in the column
// (tombstoned points are already excluded by the read path). Unlike Stats, this
// performs a full columnar scan over the column's time range, so it is O(points)
// — call it deliberately, not on a hot path. Returns 0 for an empty column.
func (c *Column) Count() (int64, error) {
	st := c.Stats()
	if !st.HasData {
		return 0, nil
	}

	var n int64
	var err error
	switch st.Scheme {
	case SchemeString, SchemeEnum:
		err = c.ScanString(st.MinMillis, st.MaxMillis, func(ts []int64, _ []string) bool {
			n += int64(len(ts))
			return true
		})
	default:
		err = c.ScanI64(st.MinMillis, st.MaxMillis, func(ts []int64, _ []int64) bool {
			n += int64(len(ts))
			return true
		})
	}
	if err != nil {
		return 0, err
	}
	return n, nil
}
