// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"errors"
	"fmt"
)

// CursorF64 reads a decimal column call by call like ScanF64, for a reader, which reads a long range in pieces and
// does something else between them, such as the callback of a script. Between two calls of Next it holds no lock; it
// keeps the reader of the chunk, in which the call before stopped, and the rest of the block it decoded. A call
// continues there, if all three hold:
//
//   - nobody took the lock of the column since the call before (colLock counts every Lock but that of a cursor),
//   - the head has no entry in the range of the call (corrections and deletions out of order change the head without
//     the lock of the column), and
//   - the call begins at the time after the last point, which the call before gave to fn; after a call, which gave
//     none, the next one begins anew.
//
// Else it begins anew like ScanF64. Either way it gives fn exactly what a fresh ScanF64 of the same range gives at this
// moment, damaged places skipped and reported like there. It loads the open block only when the sealed blocks of the
// range are through, when it continues as when it begins anew, and not, if fn stops before (R-DEC-TS-NAGO-CURSOR).
//
// Until Close it holds the buffers of a read: a window of the file up to about 256 KiB with a frame, and the decoded
// block (R-DEC-TS-NAGO-MEMORY). A CursorF64 is not safe for concurrent use; cursors of the same column are.
type CursorF64 struct {
	c   *Column
	div float64
	sc  *readScratch
	fb  *[]float64

	// valid: the state below belongs to the column as it was, when its lock counted gen
	valid bool
	gen   uint64
	// last is the time of the last point given to fn; given tells whether there was one
	last  int64
	given bool
	// ci is the chunk of sc.chunks, which is read, and inChunk whether sc.rd stands in it; open is set, when the
	// sealed blocks are through and the open block is in sc.ob, done when that is through too
	ci      int
	inChunk bool
	open    bool
	done    bool
	// bd is the decoded block, sc.bd or the open block sc.ob, and pos its first point not yet given to fn
	bd  *blockData
	pos int
}

// NewCursorF64 returns a cursor over the decimal column c. On a column of another scheme the first Next fails as
// ScanF64 does.
func (c *Column) NewCursorF64() *CursorF64 {
	return &CursorF64{c: c, div: pow10(c.schema.Decimals)}
}

// Close gives the buffers of the cursor back; afterwards the column holds no more for reading than after a query. A
// later Next begins anew.
func (cu *CursorF64) Close() {
	if cu.sc != nil {
		putScratch(cu.sc)
		cu.sc = nil
	}
	if cu.fb != nil {
		putF64(cu.fb)
		cu.fb = nil
	}
	cu.valid, cu.bd = false, nil
}

// Next calls fn like ScanF64(min, max, fn) does.
func (cu *CursorF64) Next(min, max int64, fn func(ts []int64, vals []float64) bool) error {
	c := cu.c
	if c.schema.Scheme != SchemeDecimal {
		return errSchemeMismatch
	}
	c.mu.lockQuiet()
	defer c.mu.Unlock()
	if c.dropped {
		cu.valid = false
		return errColumnGone
	}
	if err := c.head.ensure(); err != nil {
		cu.valid = false
		return err
	}
	if cu.fb == nil {
		cu.fb = f64Pool.Get().(*[]float64)
	}
	emit := func(ts, vals []int64) bool {
		if cap(*cu.fb) < len(vals) {
			*cu.fb = make([]float64, len(vals))
		}
		return fn(ts, unscaleInto((*cu.fb)[:len(vals)], vals, cu.div))
	}
	if c.head.hasRange(min, max) {
		// the slow path of a fresh read, which merges the head; the next call begins anew
		cu.valid = false
		return c.mergedRangeLocked(min, max, func(ts, vals []int64, _ []string) bool { return emit(ts, vals) })
	}
	if cu.sc == nil {
		cu.sc = getScratch()
	}
	gen := c.mu.gen.Load()
	if !cu.valid || cu.gen != gen || !cu.given || min != cu.last+1 {
		cu.begin(min)
	}
	cu.valid, cu.gen, cu.given = true, gen, false
	err := cu.read(min, max, emit)
	if err != nil {
		cu.valid = false
	}
	return err
}

// begin makes the next read begin anew at min: every chunk, whose points reach to min or later, then the open block.
// Caller holds the lock of the column.
func (cu *CursorF64) begin(min int64) {
	sc := cu.sc
	cu.c.chunksForRead(min, maxInt64, sc)
	cu.ci, cu.inChunk, cu.open, cu.done = 0, false, false, false
	sc.bd.ts, sc.bd.vals = sc.bd.ts[:0], sc.bd.vals[:0]
	cu.bd, cu.pos = &sc.bd, 0
}

// read gives fn the points of [min, max] from where the cursor stands, block by block, until fn stops or the next
// point lies after max. Caller holds the lock of the column.
func (cu *CursorF64) read(min, max int64, emit func(ts, vals []int64) bool) error {
	c, sc := cu.c, cu.sc
	for {
		if bd := cu.bd; cu.pos < len(bd.ts) {
			n := len(bd.ts)
			lo := cu.pos
			for lo < n && bd.ts[lo] < min {
				lo++
			}
			hi := lo
			for hi < n && bd.ts[hi] <= max {
				hi++
			}
			cu.pos = hi
			if hi > lo {
				cu.last, cu.given = bd.ts[hi-1], true
				if !emit(bd.ts[lo:hi], bd.vals[lo:hi]) {
					return nil
				}
			}
			if hi < n {
				return nil // the rest lies after max
			}
		}
		if cu.done {
			return nil
		}
		if cu.open {
			cu.done = true
			return nil
		}
		if !cu.inChunk {
			if cu.ci >= len(sc.chunks) {
				// the sealed blocks are through: now the open block, if it holds a point from min on
				if lo, ok := c.app.openBlockMin(); !ok || c.app.newest() < min {
					cu.open, cu.done = true, true
					return nil
				} else if lo > max {
					return nil // later, perhaps: a call with a larger max continues here
				}
				sc.ob.ts, sc.ob.vals = sc.ob.ts[:0], sc.ob.vals[:0]
				if _, err := c.app.loadOpenBlock(&sc.ob, &sc.rd); err != nil {
					return err
				}
				cu.open, cu.bd, cu.pos = true, &sc.ob, 0
				continue
			}
			ok, err := cu.enter(sc.chunks[cu.ci], min)
			if err != nil {
				return err
			}
			if !ok {
				cu.ci++
				continue
			}
			cu.inChunk = true
		}
		kind, at, frame, err := sc.rd.next(frameBlock)
		if err != nil {
			return err
		}
		if kind == frameNone {
			cu.inChunk = false
			cu.ci++
			continue
		}
		h, _ := parseBlockHeader(frame)
		if h.maxMillis < min {
			continue
		}
		ci := sc.chunks[cu.ci]
		if err := c.decodeFrame(ci, at, frame, h, &sc.bd); err != nil {
			if errors.Is(err, errUnknownFormat) {
				return err
			}
			c.reportDamaged(ci.path, at, reportBlock)
			sc.bd.ts, sc.bd.vals = sc.bd.ts[:0], sc.bd.vals[:0]
			continue
		}
		cu.bd, cu.pos = &sc.bd, 0
	}
}

// enter places the reader of the cursor at the blocks of ci from qmin on, like scanBlocks; false skips ci, a chunk
// whose header is not one of a chunk. Caller holds the lock of the column.
func (cu *CursorF64) enter(ci *chunkInfo, qmin int64) (bool, error) {
	c := cu.c
	if ci.bad {
		return false, nil
	}
	if ci.dataOff == 0 {
		head, err := readChunkHeader(c.pool, ci.path)
		if errors.Is(err, errCorruptChunk) {
			ci.bad = true
			c.reportDamaged(ci.path, 0, reportBlock)
			return false, nil
		}
		if err != nil {
			return false, err
		}
		ci.dataOff, ci.salt, ci.masked = head.dataOff, head.salt, head.masked
	}
	end := ci.sizeBytes
	if end <= 0 {
		var err error
		if end, err = statSize(c.pool, ci.path); err != nil {
			return false, fmt.Errorf("tsdb: stat chunk: %w", err)
		}
		ci.sizeBytes = end
	}
	r := &cu.sc.rd
	r.reset(c, ci.path, ci.dataOff, end, false)
	if ci.masked && end-ci.dataOff > chunkSearchFrom && ci.minMillis < qmin {
		from, err := r.search(ci.dataOff, end, qmin, ci.minMillis, ci.maxMillis)
		if err != nil {
			return false, err
		}
		r.pos, r.batch = from, readerFirstBatch
	}
	return true, nil
}
