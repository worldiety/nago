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
	"sort"
	"sync"

	"go.wdy.de/nago/pkg/ndb"
)

// Column is a handle to a single typed signal within a bucket. It owns an
// append-only head log (recent writes/updates/deletes) merged on read over a
// set of immutable sealed chunk files. All timestamps are unix milliseconds and
// each column has its own independent time axis.
//
// Column is safe for concurrent use.
type Column struct {
	db     *DB
	bucket string
	name   string
	dir    string
	schema Schema

	pool *ndb.FilePool
	head *head
	dict *enumDict // only for SchemeEnum

	mu     sync.Mutex // serializes writers and flush/compaction
	chunks []chunkInfo

	// app is the monotonic-append fast path (bounded memory, no rewrite).
	app appender
	// curMax is the largest timestamp known to the column across sealed chunks
	// and the append buffer. A write with ts > curMax takes the append fast
	// path; ts <= curMax is an out-of-order correction handled by the head.
	curMax     int64
	haveCurMax bool

	// readScratch holds reusable buffers for the read hot path. It is guarded by
	// mu (reads hold mu for the whole scan), so a single scratch per column is
	// safe and gives constant, one-time-per-read allocation.
	readScratch readScratch
}

// readScratch is the per-column reusable buffer set for reads. All fields grow
// once to the working-set size and are then reused across reads, so a scan of
// billions of points performs no per-block or per-point heap allocation.
type readScratch struct {
	fileBuf []byte    // reused chunk-file read buffer (grows to largest chunk)
	bd      blockData // reused decode target (ts/vals/strs)

	// slow-path (head overlaps range) scratch:
	headByTs    map[int64]headEntry
	emittedHead map[int64]struct{}
	outTs       []int64
	outVals     []int64
	outStrs     []string
	extra       []headEntry
}

func openColumn(db *DB, bucket, name, dir string, schema Schema) (*Column, error) {
	c := &Column{
		db:     db,
		bucket: bucket,
		name:   name,
		dir:    dir,
		schema: schema,
		pool:   db.opts.FilePool,
	}
	chunks, err := listChunks(dir)
	if err != nil {
		return nil, err
	}
	// validate headers of existing chunks; drop unreadable ones from the set.
	// A leftover pending chunk (<min>_.tsb) from a prior crash is finalized here
	// so appends resume cleanly and the tail is recovered up to the last intact
	// block (torn tail blocks are skipped by the block CRC/sync-marker scan).
	valid := chunks[:0]
	for _, ci := range chunks {
		if err := validateChunkHeader(c.pool, ci.path); err != nil {
			// leave the file on disk but skip it (house rule: keep operating)
			continue
		}
		if ci.pending {
			fci, ok := finalizeOrphanPending(c.pool, ci)
			if !ok {
				continue // empty/unreadable pending: skip
			}
			ci = fci
		}
		valid = append(valid, ci)
	}
	c.chunks = valid

	if schema.Scheme == SchemeEnum {
		d, err := openEnumDict(dir, c.pool)
		if err != nil {
			return nil, err
		}
		c.dict = d
	}

	h, err := openHead(dir, c.pool, schema.Scheme)
	if err != nil {
		return nil, err
	}
	c.head = h
	c.app.c = c

	// Recover curMax from the finalized chunk set and the head so the append
	// fast-path routing is correct after reopen. Any leftover pending chunk from
	// a prior crash is left on disk as a finalized-name-less file and will be
	// picked up by listChunks only once renamed; to keep recovery simple we scan
	// the finalized chunks (which listChunks already returned) for the max.
	for _, ci := range c.chunks {
		if !c.haveCurMax || ci.maxMillis > c.curMax {
			c.curMax = ci.maxMillis
			c.haveCurMax = true
		}
	}
	if hmax, ok := c.head.maxTS(); ok {
		if !c.haveCurMax || hmax > c.curMax {
			c.curMax = hmax
			c.haveCurMax = true
		}
	}
	return c, nil
}

// Schema returns the column's immutable schema.
func (c *Column) Schema() Schema { return c.schema }

// putNumeric writes/overwrites a numeric (decimal or enum-id) value at ts.
// Monotonic points (ts > curMax) take the constant-memory append fast path;
// out-of-order points (ts <= curMax) are corrections handled by the head.
func (c *Column) putNumeric(ts, val int64) error {
	return c.putRouted(ts, val, "")
}

// putString writes/overwrites a string value at ts (string or enum scheme).
func (c *Column) putString(ts int64, s string) error {
	if c.schema.Scheme == SchemeEnum {
		id, err := c.dict.intern(s)
		if err != nil {
			return err
		}
		return c.putRouted(ts, int64(id), "")
	}
	return c.putRouted(ts, 0, s)
}

// putRouted dispatches a write to the append fast path or the head based on
// monotonicity, holding c.mu for the whole operation.
func (c *Column) putRouted(ts, val int64, str string) error {
	c.mu.Lock()
	if !c.haveCurMax || ts > c.curMax {
		// append fast path: constant memory, no head, no rewrite.
		if err := c.app.append(ts, val, str); err != nil {
			c.mu.Unlock()
			return err
		}
		c.curMax = ts
		c.haveCurMax = true
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	// out-of-order correction / overwrite: goes to the bounded head.
	if err := c.head.put(ts, val, str); err != nil {
		return err
	}
	c.enforceHeadBound()
	return nil
}

// Delete removes the value at exactly ts (writes a tombstone that masks any
// sealed value until compaction physically removes it).
func (c *Column) Delete(ts int64) error {
	if err := c.head.tombstone(ts); err != nil {
		return err
	}
	c.enforceHeadBound()
	return nil
}

// DeleteRange removes all values with min <= ts <= max. Both the head entries
// and sealed values in the range are masked; compaction reclaims the space.
// For bulk range rewrites this tombstones the whole range in the head.
func (c *Column) DeleteRange(min, max int64) error {
	// tombstone every currently-known ts in range (head + sealed) so reads are
	// masked immediately; unknown future ts in range are not affected.
	seen := map[int64]struct{}{}
	for _, e := range c.head.snapshot(min, max) {
		seen[e.ts] = struct{}{}
	}
	c.mu.Lock()
	relevant := c.chunksForRead(min, max)
	var bd blockData
	for _, ci := range relevant {
		_ = scanChunkBlocks(c.pool, ci.path, func(frame []byte, h blockHeader) bool {
			if h.maxMillis < min || h.minMillis > max {
				return true
			}
			if decodeBlockBody(frame, h, c.schema.Scheme, &bd) != nil {
				return true
			}
			for _, t := range bd.ts {
				if t >= min && t <= max {
					seen[t] = struct{}{}
				}
			}
			return true
		})
	}
	// also tombstone in-memory buffered points in range (not yet on disk)
	c.app.snapshotBuffered(min, max, func(ts, _ []int64, _ []string) bool {
		for _, t := range ts {
			seen[t] = struct{}{}
		}
		return true
	})
	c.mu.Unlock()
	for t := range seen {
		if err := c.head.tombstone(t); err != nil {
			return err
		}
	}
	c.enforceHeadBound()
	return nil
}

// enforceHeadBound keeps the in-memory head within the configured hard cap. When
// the head grows past MaxHeadPoints it flushes synchronously (blocking the
// writer briefly) so that RAM can never exceed the cap regardless of how
// out-of-order the workload is. Below the cap it enqueues async compaction so
// the common case stays off the write path.
func (c *Column) enforceHeadBound() {
	n := c.head.len()
	if n >= c.db.opts.MaxHeadPoints {
		c.mu.Lock()
		// re-check under lock; a concurrent flush may have drained it already.
		if c.head.len() >= c.db.opts.MaxHeadPoints {
			_ = c.flushLocked()
		}
		c.mu.Unlock()
		return
	}
	if c.needsFlush() {
		c.db.compactor.enqueue(c)
	}
}

// needsFlush reports whether the head has grown enough to warrant folding into
// a chunk, or accumulated enough tombstones/overwrites to warrant compaction.
func (c *Column) needsFlush() bool {
	c.head.mu.RLock()
	bytes := c.head.bytesLogged
	tomb := c.head.tombstones
	n := c.head.tree.Len()
	c.head.mu.RUnlock()
	if bytes >= c.db.opts.FlushBytes {
		return true
	}
	if n > 0 && float64(tomb)/float64(n) >= c.db.opts.CompactTombstoneRatio {
		return true
	}
	return false
}

func (c *Column) chunksOverlapping(min, max int64) []chunkInfo {
	var out []chunkInfo
	for _, ci := range c.chunks {
		if ci.maxMillis < min || ci.minMillis > max {
			continue
		}
		out = append(out, ci)
	}
	return out
}

// chunksForRead returns the finalized overlapping chunks plus, appended last,
// the appender's still-open pending chunk when it overlaps the range. The
// pending chunk holds monotonic blocks already sealed to disk but not yet
// finalized; reads must include it so data written during a large burst is
// visible before an explicit Flush. Its blocks are the newest on disk, so it is
// scanned after the finalized chunks; the in-memory append buffer (newer still)
// is emitted separately by the caller.
func (c *Column) chunksForRead(min, max int64) []chunkInfo {
	out := c.chunksOverlapping(min, max)
	if pc, ok := c.app.pendingChunk(); ok {
		if !(pc.maxMillis < min || pc.minMillis > max) {
			out = append(out, pc)
		}
	}
	return out
}

// mergedRange yields every non-deleted point with min <= ts <= max in ascending
// timestamp order, merging sealed chunks with the head (head wins). It decodes
// block by block and calls fn with parallel slices; fn returns false to stop.
// For enum columns vals carries dictionary ids; string resolution is the
// caller's job. For string columns strs is populated and vals is nil.
//
// This is the single read primitive shared by the public API and compaction.
//
// Allocation profile: at most a constant, one-time set of scratch buffers per
// call (the reusable file-read buffer plus, only when the head overlaps the
// range, the override maps). The common post-flush read (empty head) takes a
// fast path that performs zero heap allocation beyond growing the single file
// buffer once, and streams decoded blocks straight to fn without copying.
func (c *Column) mergedRange(min, max int64, fn func(ts []int64, vals []int64, strs []string) bool) error {
	// Hold c.mu for the whole read so chunk files cannot be removed by a
	// concurrent flush/compaction while we scan them. Compaction is off the
	// write path and infrequent; correctness is prioritized here. A future
	// optimization can replace this with per-chunk refcounting for lock-free
	// reads over immutable files.
	c.mu.Lock()
	defer c.mu.Unlock()
	chunks := c.chunksForRead(min, max)

	sc := &c.readScratch
	numeric := c.schema.Scheme != SchemeString

	// Fast path: the head does not overlap the query range. This is the
	// dominant read case after a flush. We stream each decoded block directly
	// to fn with no per-point copy and no maps.
	if !c.head.hasRange(min, max) {
		if err := c.streamSealed(min, max, chunks, numeric, sc, fn); err != nil {
			return err
		}
		// include the newest monotonic tail still buffered in the appender.
		c.app.snapshotBuffered(min, max, fn)
		return nil
	}

	// Slow path: the head overlaps; fall back to the map-based merge.
	headEntries := c.head.snapshot(min, max)
	// Head entries whose timestamp lies within the buffered (unsealed) range are
	// overrides of buffered points and are emitted by the buffered pass below, so
	// exclude them from emitMerged's head-only emission to avoid duplicates.
	bufMin, bufOK := c.app.bufferedMinTS()
	emitMax := max
	if bufOK && bufMin-1 < emitMax {
		emitMax = bufMin - 1
	}
	if err := c.emitMerged(min, emitMax, chunks, headEntries, numeric, sc, fn); err != nil {
		return err
	}
	// The buffered (unsealed) tail may be overridden or tombstoned by head
	// entries (an overwrite/delete of a just-appended point). Apply those.
	c.snapshotBufferedMasked(min, max, headEntries, numeric, fn)
	return nil
}

// snapshotBufferedMasked emits the appender's buffered points in [min,max],
// applying any head override/tombstone that targets a buffered timestamp.
func (c *Column) snapshotBufferedMasked(min, max int64, headEntries []headEntry,
	numeric bool, fn func(ts, vals []int64, strs []string) bool) {

	if c.app.bufferedCount() == 0 {
		return
	}
	// small linear apply: buffered set is bounded by one block; head slice is
	// bounded by the query. Build a tiny override lookup only over the buffered
	// range to stay allocation-light.
	c.app.snapshotBufferedFilter(min, max, func(ts int64) (int64, string, bool, bool) {
		for i := range headEntries {
			if headEntries[i].ts == ts {
				e := headEntries[i]
				return e.val, e.str, e.tombstone, true
			}
		}
		return 0, "", false, false
	}, numeric, fn)
}

// streamSealed emits sealed points in [min,max] with no head overrides. Each
// decoded block is passed straight to fn; the only trimming is when a block
// straddles the query bounds (rare, at the two ends of the range).
func (c *Column) streamSealed(min, max int64, chunks []chunkInfo, numeric bool,
	sc *readScratch, fn func(ts []int64, vals []int64, strs []string) bool) error {

	scheme := c.schema.Scheme
	bd := &sc.bd
	stop := false
	for _, ci := range chunks {
		if stop {
			break
		}
		err := scanChunkBlocksBuf(c.pool, ci.path, ci.sizeBytes, &sc.fileBuf, func(frame []byte, h blockHeader) bool {
			if h.maxMillis < min || h.minMillis > max {
				return true
			}
			if decodeBlockBody(frame, h, scheme, bd) != nil {
				return true // skip corrupt block
			}
			ts := bd.ts
			// If the whole block is inside the query bounds (the common case),
			// hand the decoded slices to fn without any copy.
			if ts[0] >= min && ts[len(ts)-1] <= max {
				if !fn(ts, sealedVals(bd, numeric), sealedStrs(bd, numeric)) {
					stop = true
					return false
				}
				return true
			}
			// Partial-overlap block: trim to [min,max] into reusable out slices.
			lo := 0
			for lo < len(ts) && ts[lo] < min {
				lo++
			}
			hi := len(ts)
			for hi > lo && ts[hi-1] > max {
				hi--
			}
			if lo >= hi {
				return true
			}
			if numeric {
				if !fn(ts[lo:hi], sealedVals(bd, true)[lo:hi], nil) {
					stop = true
					return false
				}
			} else {
				if !fn(ts[lo:hi], nil, bd.strs[lo:hi]) {
					stop = true
					return false
				}
			}
			return true
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// sealedVals returns the numeric value slice for the decoded block. Enum ids
// and decimal values share the same int64 vals slice, so no conversion is
// needed. Returns nil for the string scheme.
func sealedVals(bd *blockData, numeric bool) []int64 {
	if !numeric {
		return nil
	}
	return bd.vals
}

func sealedStrs(bd *blockData, numeric bool) []string {
	if numeric {
		return nil
	}
	return bd.strs
}

func sealedVal(bd *blockData, i int) int64 {
	return bd.vals[i]
}

// emitMerged is the slow path used only when the head overlaps the query range.
// It applies head overrides and tombstones over the sealed blocks. All scratch
// (override map, seen-set, out buffers) is reused from sc, so repeated reads do
// not re-allocate. Correctness matches the fast path; only the head-overlap
// case pays for the maps.
func (c *Column) emitMerged(min, max int64, chunks []chunkInfo, headEntries []headEntry,
	numeric bool, sc *readScratch, fn func(ts []int64, vals []int64, strs []string) bool) error {

	scheme := c.schema.Scheme

	if sc.headByTs == nil {
		sc.headByTs = make(map[int64]headEntry, len(headEntries))
	} else {
		clear(sc.headByTs)
	}
	if sc.emittedHead == nil {
		sc.emittedHead = make(map[int64]struct{}, len(headEntries))
	} else {
		clear(sc.emittedHead)
	}
	headByTs := sc.headByTs
	emittedHead := sc.emittedHead
	for _, e := range headEntries {
		headByTs[e.ts] = e
	}

	bd := &sc.bd
	outTs := sc.outTs[:0]
	outVals := sc.outVals[:0]
	outStrs := sc.outStrs[:0]

	flush := func() bool {
		if len(outTs) == 0 {
			return true
		}
		var vals []int64
		var strs []string
		if numeric {
			vals = outVals
		} else {
			strs = outStrs
		}
		ok := fn(outTs, vals, strs)
		outTs = outTs[:0]
		outVals = outVals[:0]
		outStrs = outStrs[:0]
		return ok
	}

	stop := false
	for _, ci := range chunks {
		if stop {
			break
		}
		err := scanChunkBlocksBuf(c.pool, ci.path, ci.sizeBytes, &sc.fileBuf, func(frame []byte, h blockHeader) bool {
			if h.maxMillis < min || h.minMillis > max {
				return true
			}
			if decodeBlockBody(frame, h, scheme, bd) != nil {
				return true
			}
			for i, t := range bd.ts {
				if t < min || t > max {
					continue
				}
				if he, ok := headByTs[t]; ok {
					emittedHead[t] = struct{}{}
					if he.tombstone {
						continue
					}
					outTs = append(outTs, t)
					if numeric {
						outVals = append(outVals, he.val)
					} else {
						outStrs = append(outStrs, he.str)
					}
					continue
				}
				outTs = append(outTs, t)
				if numeric {
					outVals = append(outVals, sealedVal(bd, i))
				} else {
					outStrs = append(outStrs, bd.strs[i])
				}
			}
			if !flush() {
				stop = true
				return false
			}
			return true
		})
		if err != nil {
			sc.outTs, sc.outVals, sc.outStrs = outTs, outVals, outStrs
			return err
		}
	}
	if stop {
		sc.outTs, sc.outVals, sc.outStrs = outTs, outVals, outStrs
		return nil
	}

	// emit head-only entries (not present in any sealed block) in ascending
	// order, bounded by max (buffered-range overrides are emitted separately).
	extra := sc.extra[:0]
	for _, e := range headEntries {
		if e.ts > max {
			continue
		}
		if _, done := emittedHead[e.ts]; done || e.tombstone {
			continue
		}
		extra = append(extra, e)
	}
	sort.Slice(extra, func(i, j int) bool { return extra[i].ts < extra[j].ts })
	for _, e := range extra {
		outTs = append(outTs, e.ts)
		if numeric {
			outVals = append(outVals, e.val)
		} else {
			outStrs = append(outStrs, e.str)
		}
	}
	flush()
	sc.outTs, sc.outVals, sc.outStrs, sc.extra = outTs, outVals, outStrs, extra
	return nil
}

// flushLocked makes all buffered/append data durable and folds the out-of-order
// head into sealed chunks. Caller holds c.mu. Two parts:
//
//  1. The append fast-path pending chunk is finalized so its tail is on disk.
//  2. If the head holds out-of-order corrections, the overlapping chunks are
//     rewritten to physically apply overwrites/tombstones, then the head resets.
func (c *Column) flushLocked() error {
	// Part 1: finalize the monotonic append pending chunk (cheap, no rewrite).
	if err := c.app.flush(); err != nil {
		return err
	}

	// Part 2: fold the out-of-order head, if any.
	headEntries := c.head.snapshot(minInt64, maxInt64)
	if len(headEntries) == 0 {
		return nil
	}

	// Determine the time span the head touches, then rewrite the union of that
	// span with any overlapping sealed chunks into fresh chunk files.
	var hmin, hmax int64 = maxInt64, minInt64
	for _, e := range headEntries {
		if e.ts < hmin {
			hmin = e.ts
		}
		if e.ts > hmax {
			hmax = e.ts
		}
	}

	overlap := c.chunksOverlapping(hmin, hmax)

	// Expand the rewrite window to the full extent of the overlapping chunks so
	// that points inside those chunks but outside the head's span are carried
	// over rather than dropped when the old chunk files are replaced.
	for _, ci := range overlap {
		if ci.minMillis < hmin {
			hmin = ci.minMillis
		}
		if ci.maxMillis > hmax {
			hmax = ci.maxMillis
		}
	}

	// Build the merged, ordered, de-duplicated point set across overlapping
	// sealed chunks + head (head wins, tombstones drop).
	merged := map[int64]mergedPoint{}
	var bd blockData
	for _, ci := range overlap {
		if err := scanChunkBlocks(c.pool, ci.path, func(frame []byte, h blockHeader) bool {
			if decodeBlockBody(frame, h, c.schema.Scheme, &bd) != nil {
				return true
			}
			for i, t := range bd.ts {
				merged[t] = sealedPoint(&bd, c.schema.Scheme, i)
			}
			return true
		}); err != nil {
			return err
		}
	}
	for _, e := range headEntries {
		if e.tombstone {
			delete(merged, e.ts)
			continue
		}
		merged[e.ts] = mergedPoint{val: e.val, str: e.str}
	}

	// also carry sealed chunks that are fully outside [hmin,hmax] untouched
	order := make([]int64, 0, len(merged))
	for t := range merged {
		order = append(order, t)
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })

	// Remove the old overlapping chunk files BEFORE writing the replacements.
	// A replacement may legitimately reuse the same <min>_<max>.tsb name (e.g.
	// an in-place overwrite that does not change the range), so writing first
	// and deleting after would delete the fresh file.
	for _, ci := range overlap {
		c.pool.Evict(ci.path)
		if err := os.Remove(ci.path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("tsdb: remove old chunk: %w", err)
		}
	}

	// write new chunk files for the merged range, respecting the split policy
	newChunks, err := c.writeChunks(order, merged)
	if err != nil {
		return err
	}

	kept := make([]chunkInfo, 0, len(c.chunks))
	for _, ci := range c.chunks {
		if ci.maxMillis < hmin || ci.minMillis > hmax {
			kept = append(kept, ci)
		}
	}
	c.chunks = append(kept, newChunks...)
	sort.Slice(c.chunks, func(i, j int) bool { return c.chunks[i].minMillis < c.chunks[j].minMillis })

	return c.head.reset()
}

type mergedPoint struct {
	val int64
	str string
}

func sealedPoint(bd *blockData, s Scheme, i int) mergedPoint {
	if s == SchemeString {
		return mergedPoint{str: bd.strs[i]}
	}
	return mergedPoint{val: sealedVal(bd, i)}
}

// writeChunks writes the ordered points into new chunk files, sealing per the
// split policy, and returns the created chunkInfos.
func (c *Column) writeChunks(order []int64, merged map[int64]mergedPoint) ([]chunkInfo, error) {
	var out []chunkInfo
	i := 0
	for i < len(order) {
		startTs := order[i]
		name := pendingChunkName(startTs)
		path := filepath.Join(c.dir, name)
		if err := writeChunkHeader(c.pool, path); err != nil {
			return nil, err
		}
		var fileOff int64 = chunkHeaderLen
		var points int64
		var lastTs int64 = startTs

		var bd blockData
		frameBuf := make([]byte, 0, 64<<10)

		flushBlock := func() error {
			if bd.count() == 0 {
				return nil
			}
			frameBuf = frameBuf[:0]
			f, err := encodeBlock(frameBuf, &bd, c.schema.Scheme, c.db.opts.Compress)
			if err != nil {
				return err
			}
			if _, err := c.pool.WriteAt(path, f, fileOff); err != nil {
				return err
			}
			fileOff += int64(len(f))
			bd.ts = bd.ts[:0]
			bd.vals = bd.vals[:0]
			bd.strs = bd.strs[:0]
			return nil
		}

		sealed := false
		for i < len(order) {
			ts := order[i]
			stats := ChunkStats{
				SizeBytes: fileOff, Points: points,
				MinMillis: startTs, MaxMillis: lastTs, NextMillis: ts,
			}
			if points > 0 && c.db.opts.Split(stats) {
				sealed = true
				break
			}
			mp := merged[ts]
			bd.ts = append(bd.ts, ts)
			switch c.schema.Scheme {
			case SchemeDecimal, SchemeEnum:
				bd.vals = append(bd.vals, mp.val)
			case SchemeString:
				bd.strs = append(bd.strs, mp.str)
			}
			lastTs = ts
			points++
			i++
			if bd.count() >= c.db.opts.BlockPoints {
				if err := flushBlock(); err != nil {
					return nil, err
				}
			}
		}
		if err := flushBlock(); err != nil {
			return nil, err
		}

		// finalize: rename pending -> final with true min/max
		finalName := finalChunkName(startTs, lastTs)
		finalPath := filepath.Join(c.dir, finalName)
		c.pool.Evict(path)
		if err := os.Rename(path, finalPath); err != nil {
			return nil, fmt.Errorf("tsdb: finalize chunk: %w", err)
		}
		out = append(out, chunkInfo{path: finalPath, minMillis: startTs, maxMillis: lastTs, sizeBytes: fileOff})
		_ = sealed
	}
	return out, nil
}

func (c *Column) close() error {
	if c.head != nil {
		return c.head.close()
	}
	return nil
}

const (
	minInt64 int64 = -1 << 63
	maxInt64 int64 = 1<<63 - 1
)
