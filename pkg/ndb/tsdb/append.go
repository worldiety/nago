// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"fmt"
	"os"
	"path/filepath"
)

// appender is the monotonic-append fast path of a column. Points whose
// timestamp is strictly greater than everything already stored bypass the
// in-memory head entirely and stream directly into sealed chunk files:
//
//   - Incoming points accumulate in an in-memory block buffer (blockData).
//   - When the buffer reaches BlockPoints, it is encoded and appended to the
//     current pending .tsb chunk file (one pwrite), and the buffer is cleared.
//   - When the split policy fires, the pending chunk is finalized (rename to its
//     <min>_<max>.tsb name) and a new pending chunk begins.
//
// Memory is therefore bounded to a single block per column regardless of how
// many points are ingested — a 20-billion-point burst runs in constant memory
// with no btree and no chunk rewrite. This mirrors the msgstore ingest model.
//
// Durability follows the no-fsync house rule: the not-yet-sealed block lives in
// memory and is lost on a hard crash; sealed blocks and finalized chunks are
// durable. Callers needing the tail on disk call Column.Flush.
type appender struct {
	c *Column

	buf blockData // in-memory accumulation block

	pendingPath string // path of the open pending chunk, "" if none
	pendingMin  int64  // min ts of the pending chunk
	pendingMax  int64  // max ts written so far to the pending chunk
	pendingOff  int64  // current append offset in the pending chunk file
	pendingPts  int64  // points written to the pending chunk (sealed blocks only)

	frameScratch []byte // reusable encode buffer
}

// bufferedCount returns the number of points currently in the in-memory block
// buffer (not yet appended to the pending chunk file).
func (a *appender) bufferedCount() int { return a.buf.count() }

// bufferedMinTS returns the smallest timestamp in the in-memory buffer, if any.
func (a *appender) bufferedMinTS() (int64, bool) {
	if a.buf.count() == 0 {
		return 0, false
	}
	return a.buf.ts[0], true
}

// append adds one monotonic point (ts strictly greater than the column max).
// scheme-specific value is placed via the provided setter closure semantics
// inlined here to avoid an interface. Caller holds c.mu.
func (a *appender) append(ts, val int64, str string) error {
	// split decision is evaluated against the pending chunk before adding.
	if a.pendingPath != "" && a.buf.count() == 0 {
		stats := ChunkStats{
			SizeBytes:  a.pendingOff,
			Points:     a.pendingPts,
			MinMillis:  a.pendingMin,
			MaxMillis:  a.pendingMax,
			NextMillis: ts,
		}
		if a.pendingPts > 0 && a.c.db.opts.Split(stats) {
			if err := a.finalizePending(); err != nil {
				return err
			}
		}
	}

	if a.pendingPath == "" {
		if err := a.openPending(ts); err != nil {
			return err
		}
	}

	a.buf.ts = append(a.buf.ts, ts)
	switch a.c.schema.Scheme {
	case SchemeDecimal, SchemeEnum:
		a.buf.vals = append(a.buf.vals, val)
	case SchemeString:
		a.buf.strs = append(a.buf.strs, str)
	}
	if ts > a.pendingMax {
		a.pendingMax = ts
	}

	if a.buf.count() >= a.c.db.opts.BlockPoints {
		return a.sealBlock()
	}
	return nil
}

func (a *appender) openPending(minTs int64) error {
	name := pendingChunkName(minTs)
	path := filepath.Join(a.c.dir, name)
	if err := writeChunkHeader(a.c.pool, path); err != nil {
		return err
	}
	a.pendingPath = path
	a.pendingMin = minTs
	a.pendingMax = minTs
	a.pendingOff = chunkHeaderLen
	a.pendingPts = 0
	return nil
}

// sealBlock encodes the in-memory buffer into a block and appends it to the
// pending chunk file, then clears the buffer.
func (a *appender) sealBlock() error {
	if a.buf.count() == 0 {
		return nil
	}
	a.frameScratch = a.frameScratch[:0]
	frame, err := encodeBlock(a.frameScratch, &a.buf, a.c.schema.Scheme, a.c.db.opts.Compress)
	if err != nil {
		return err
	}
	a.frameScratch = frame
	if _, err := a.c.pool.WriteAt(a.pendingPath, frame, a.pendingOff); err != nil {
		return fmt.Errorf("tsdb: append block: %w", err)
	}
	a.pendingOff += int64(len(frame))
	a.pendingPts += int64(a.buf.count())
	a.buf.ts = a.buf.ts[:0]
	a.buf.vals = a.buf.vals[:0]
	a.buf.strs = a.buf.strs[:0]
	return nil
}

// finalizePending seals any buffered block and renames the pending chunk to its
// final <min>_<max>.tsb name, registering it in the column's chunk set.
func (a *appender) finalizePending() error {
	if a.pendingPath == "" {
		return nil
	}
	if err := a.sealBlock(); err != nil {
		return err
	}
	if a.pendingPts == 0 {
		// empty pending chunk (never happens in practice) — remove it
		a.c.pool.Evict(a.pendingPath)
		_ = os.Remove(a.pendingPath)
		a.pendingPath = ""
		return nil
	}
	finalPath := filepath.Join(a.c.dir, finalChunkName(a.pendingMin, a.pendingMax))
	a.c.pool.Evict(a.pendingPath)
	if err := os.Rename(a.pendingPath, finalPath); err != nil {
		return fmt.Errorf("tsdb: finalize appended chunk: %w", err)
	}
	a.c.chunks = append(a.c.chunks, chunkInfo{
		path:      finalPath,
		minMillis: a.pendingMin,
		maxMillis: a.pendingMax,
		sizeBytes: a.pendingOff,
	})
	a.pendingPath = ""
	return nil
}

// flush finalizes the pending chunk so all appended data is durable on disk.
// Caller holds c.mu.
func (a *appender) flush() error {
	return a.finalizePending()
}

// snapshotBuffered appends the in-memory (not yet sealed) buffered points that
// fall in [min,max] to the read output via the provided callback. Used by the
// read path to include the newest tail that has not been written to a chunk yet.
func (a *appender) snapshotBuffered(min, max int64, fn func(ts, vals []int64, strs []string) bool) bool {
	n := a.buf.count()
	if n == 0 {
		return true
	}
	// buffer is monotonic ascending; find the in-range window.
	lo := 0
	for lo < n && a.buf.ts[lo] < min {
		lo++
	}
	hi := n
	for hi > lo && a.buf.ts[hi-1] > max {
		hi--
	}
	if lo >= hi {
		return true
	}
	if a.c.schema.Scheme == SchemeString {
		return fn(a.buf.ts[lo:hi], nil, a.buf.strs[lo:hi])
	}
	return fn(a.buf.ts[lo:hi], sealedVals(&a.buf, true)[lo:hi], nil)
}

// snapshotBufferedFilter emits buffered points in [min,max], consulting override
// for each timestamp: override returns (val, str, tombstone, present). Present
// tombstones drop the point; present non-tombstones replace its value. This is
// used only on the slow read path where the head may target a buffered (unsealed)
// point, so per-point emission is acceptable (buffer is bounded to one block).
func (a *appender) snapshotBufferedFilter(min, max int64,
	override func(ts int64) (val int64, str string, tomb bool, present bool),
	numeric bool, fn func(ts, vals []int64, strs []string) bool) {

	n := a.buf.count()
	if n == 0 {
		return
	}
	var ts [1]int64
	var vals [1]int64
	var strs [1]string
	for i := 0; i < n; i++ {
		t := a.buf.ts[i]
		if t < min || t > max {
			continue
		}
		if ov, os, tomb, present := override(t); present {
			if tomb {
				continue
			}
			ts[0] = t
			if numeric {
				vals[0] = ov
				if !fn(ts[:], vals[:], nil) {
					return
				}
			} else {
				strs[0] = os
				if !fn(ts[:], nil, strs[:]) {
					return
				}
			}
			continue
		}
		ts[0] = t
		if numeric {
			vals[0] = sealedVal(&a.buf, i)
			if !fn(ts[:], vals[:], nil) {
				return
			}
		} else {
			strs[0] = a.buf.strs[i]
			if !fn(ts[:], nil, strs[:]) {
				return
			}
		}
	}
}
