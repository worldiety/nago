// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import "time"

// ChunkStats is the state of the current pending chunk passed to a SplitFunc so
// it can decide whether to seal it and start a new one.
type ChunkStats struct {
	// SizeBytes is the current on-disk size of the pending chunk.
	SizeBytes int64
	// Points is the number of points written to the pending chunk so far.
	Points int64
	// MinMillis / MaxMillis are the time bounds of the pending chunk.
	MinMillis int64
	MaxMillis int64
	// NextMillis is the timestamp of the point about to be written.
	NextMillis int64
}

// SplitFunc decides, before a point is appended to the pending chunk, whether
// that chunk should first be sealed (finalized) so the point starts a new one.
// It is the sole partitioning policy: the on-disk format itself is timezone
// free (chunks are named by raw epoch millis), so any calendar interpretation
// lives entirely inside the SplitFunc.
type SplitFunc func(s ChunkStats) bool

const defaultMaxChunkBytes int64 = 64 << 20 // 64 MiB

// SplitBySize seals the chunk once it reaches maxBytes.
func SplitBySize(maxBytes int64) SplitFunc {
	return func(s ChunkStats) bool { return s.SizeBytes >= maxBytes }
}

// SplitByCount seals the chunk once it reaches maxPoints.
func SplitByCount(maxPoints int64) SplitFunc {
	return func(s ChunkStats) bool { return s.Points >= maxPoints }
}

// CombineSplits seals when any of the given policies wants to seal.
func CombineSplits(fns ...SplitFunc) SplitFunc {
	return func(s ChunkStats) bool {
		for _, fn := range fns {
			if fn(s) {
				return true
			}
		}
		return false
	}
}

// SplitByQuarter seals when the point about to be written falls into a later
// calendar quarter than the pending chunk's start, interpreted in loc. This is
// timezone-aware by design; the storage layer never sees loc.
func SplitByQuarter(loc *time.Location) SplitFunc {
	return func(s ChunkStats) bool {
		if s.Points == 0 {
			return false
		}
		start := time.UnixMilli(s.MinMillis).In(loc)
		next := time.UnixMilli(s.NextMillis).In(loc)
		if next.Year() != start.Year() {
			return true
		}
		return quarter(next.Month()) != quarter(start.Month())
	}
}

func quarter(m time.Month) int { return (int(m) - 1) / 3 }

// defaultSplit is the default partitioning policy: seal on a German-time
// calendar-quarter boundary OR at 64 MiB, whichever comes first. Falls back to
// UTC if the Europe/Berlin zone cannot be loaded.
func defaultSplit() SplitFunc {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		loc = time.UTC
	}
	return CombineSplits(SplitByQuarter(loc), SplitBySize(defaultMaxChunkBytes))
}
