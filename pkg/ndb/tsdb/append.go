// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"fmt"
	"path/filepath"
)

// appender is the monotonic-append fast path of a column. Points whose
// timestamp is strictly greater than everything already stored bypass the
// in-memory head entirely and stream directly into chunk files:
//
//   - Incoming points form the open block. Those not yet saved are held encoded in memory (pointEnc, a few bytes per
//     point).
//   - Column.Sync appends them as a continuation record behind the sealed blocks of the open chunk (one pwrite) and
//     forgets them; the open block goes on with the next point.
//   - When the open block reaches BlockPoints, its saved points from the records and those in memory are sealed as one
//     block, which takes the place of the records (see sealBlock), so that a block is the same with or without Sync.
//   - When the split policy fires, the open chunk is finalized (rename to its
//     <min>_<max>.tsb name) and a new open chunk begins.
//
// Memory is therefore bounded to the unsaved points of one block per column regardless of how many points are
// ingested, and a Sync or a sealed block leaves none of it. This mirrors the msgstore ingest model.
//
// Durability follows the no-fsync house rule: the unsaved points live in memory and are lost on a crash of the
// process; records, sealed blocks and finalized chunks are written. DB.Close seals the open block, and the next open of
// the column continues the open chunk (see adopt), so only the split policy ends a chunk of a monotonic column,
// however often the process restarts.
type appender struct {
	c *Column

	pendingPath string // path of the open pending chunk, "" if none
	pendingMin  int64  // min ts of the pending chunk, from its name
	dataOff     int64  // where its blocks begin
	salt        uint64 // the key of its masks
	masked      bool   // version 2: bodies are masked
	slotSeq     uint32 // the newest slot of its header
	blocksEnd   int64  // end of its sealed blocks
	blocksPts   int64  // points of its sealed blocks
	blocksMax   int64  // newest time of its sealed blocks, valid if blocksPts > 0
	fileEnd     int64  // its size: the sealed blocks and the continuation records behind them
	savedPts    int64  // points of the open block in those records
	savedMin    int64
	savedMax    int64

	mem pointEnc // the points of the open block, which are not saved yet
}

// unsynced is the number of points only the memory holds.
func (a *appender) unsynced() int { return a.mem.n }

// blockPoints is the number of points of the open block, saved or not.
func (a *appender) blockPoints() int64 { return a.savedPts + int64(a.mem.n) }

// newest returns the newest time of the open chunk, in memory or on disk.
func (a *appender) newest() int64 {
	switch {
	case a.mem.n > 0:
		return a.mem.maxTs
	case a.savedPts > 0:
		return a.savedMax
	}
	return a.blocksMax
}

// openBlockMin returns the oldest time of the open block, if it has a point.
func (a *appender) openBlockMin() (int64, bool) {
	switch {
	case a.savedPts > 0:
		return a.savedMin, true
	case a.mem.n > 0:
		return a.mem.minTs, true
	}
	return 0, false
}

// append adds one monotonic point (ts strictly greater than the column max). Caller holds c.mu.
func (a *appender) append(ts, val int64, str string) error {
	// The split policy is asked at the first point after a Sync, after the open and at a block boundary: whenever no
	// point waits in memory.
	if a.pendingPath != "" && a.mem.n == 0 {
		stats := ChunkStats{
			SizeBytes:  a.fileEnd,
			Points:     a.blocksPts + a.savedPts,
			MinMillis:  a.pendingMin,
			MaxMillis:  a.newest(),
			NextMillis: ts,
		}
		if stats.Points > 0 && a.c.db.opts.Split(stats) {
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

	a.mem.add(a.c.schema.Scheme, ts, val, str)
	if a.blockPoints() >= int64(a.c.db.opts.BlockPoints) {
		return a.sealBlock()
	}
	return nil
}

func (a *appender) openPending(minTs int64) error {
	path := filepath.Join(a.c.dir, pendingChunkName(minTs))
	slot := chunkSlot{seq: 1, end: chunkDataOff}
	salt := newSalt()
	if _, err := a.c.pool.WriteAt(path, chunkHeader(slot, salt), 0); err != nil {
		return fmt.Errorf("tsdb: write chunk header: %w", err)
	}
	// a file of that name, which was no chunk, has nothing behind the header
	if err := a.c.pool.Truncate(path, chunkDataOff); err != nil {
		return fmt.Errorf("tsdb: write chunk header: %w", err)
	}
	a.adopt(openState{path: path, minMillis: minTs, dataOff: chunkDataOff, salt: salt, masked: true, slotSeq: slot.seq,
		blocksEnd: chunkDataOff, fileEnd: chunkDataOff})
	return nil
}

// adopt makes the open chunk of st the pending chunk, so that the next appends continue it instead of beginning a new
// file. The split policy is asked before the next point, as after a Sync. Used for the open chunk of an earlier life
// and for the open rest of a rewrite by a compaction. Caller holds c.mu or is opening the column.
func (a *appender) adopt(st openState) {
	a.pendingPath = st.path
	a.pendingMin = st.minMillis
	a.dataOff = st.dataOff
	a.salt = st.salt
	a.masked = st.masked
	a.slotSeq = st.slotSeq
	a.blocksEnd = st.blocksEnd
	a.blocksPts = st.blocksPts
	a.blocksMax = st.blocksMax
	a.fileEnd = st.fileEnd
	a.savedPts = st.savedPts
	a.savedMin = st.savedMin
	a.savedMax = st.savedMax
	a.mem.release()
}

// forgetPending drops the state of the pending chunk and the points in memory,
// whose file the caller has removed. Caller holds c.mu.
func (a *appender) forgetPending() {
	*a = appender{c: a.c}
}

// sync appends the points in memory as a continuation record behind the open chunk and forgets them. It writes one
// frame into one file and changes nothing else, so that a crash in its middle costs only these points: the next open
// cuts the torn record off. Caller holds c.mu.
func (a *appender) sync() error {
	if a.mem.n == 0 {
		return nil
	}
	rec := appendRecord(nil, recordDelta, a.mem.n, a.mem.buf, a.masked, a.salt, a.fileEnd)
	if _, err := a.c.pool.WriteAt(a.pendingPath, rec, a.fileEnd); err != nil {
		return fmt.Errorf("tsdb: append continuation record: %w", err)
	}
	a.fileEnd += int64(len(rec))
	if a.savedPts == 0 {
		a.savedMin = a.mem.minTs
	}
	a.savedPts += int64(a.mem.n)
	a.savedMax = a.mem.maxTs
	a.mem.release()
	return nil
}

// loadOpenBlock appends the points of the open block, the saved ones from its records and those in memory, to bd in
// ascending order. Caller holds c.mu.
func (a *appender) loadOpenBlock(bd *blockData, r *frameReader) error {
	if a.savedPts > 0 {
		r.reset(a.c.pool, a.pendingPath, a.blocksEnd, a.fileEnd)
		for {
			kind, at, frame, err := r.next(frameRecord)
			if err != nil {
				return err
			}
			if kind == frameNone {
				break
			}
			// a record of a later version failed the open of the column already (recoverOpenChunk)
			_, count, body, _ := parseRecord(frame, len(frame), a.salt, at)
			n := len(bd.ts)
			if !decodePoints(body, count, a.c.schema.Scheme, bd) {
				bd.ts, bd.vals, bd.strs = bd.ts[:n], bd.vals[:min(n, len(bd.vals))], bd.strs[:min(n, len(bd.strs))]
			}
		}
	}
	if a.mem.n > 0 && !decodePoints(a.mem.buf, a.mem.n, a.c.schema.Scheme, bd) {
		return fmt.Errorf("tsdb: open block in memory: %w", errCorruptBlock)
	}
	return nil
}

// sealBlock seals the open block, its saved and its unsaved points, as one block at the end of the sealed blocks and
// clears the memory. Without saved points the block is just appended. With them, the block takes the place of their
// records: a copy record of the saved points first goes behind both, then the block is written over the records, then
// the file is cut behind the block. A crash at any byte of that leaves the saved points readable, as the records, the
// copy or the block (see recoverOpenChunk), and the file is the same as if no Sync had run. Then the newer slot of the
// header says where the blocks end. Caller holds c.mu.
func (a *appender) sealBlock() error {
	if a.blockPoints() == 0 {
		return nil
	}
	sc := getScratch()
	defer putScratch(sc)
	bd := &sc.ob
	if err := a.loadOpenBlock(bd, &sc.rd); err != nil {
		return err
	}
	saved := int(a.savedPts)
	if len(bd.ts) < saved {
		saved = len(bd.ts) // a record, which can not be read any more, is gone
	}
	if len(bd.ts) == 0 {
		if err := a.c.pool.Truncate(a.pendingPath, a.blocksEnd); err != nil {
			return fmt.Errorf("tsdb: cut unreadable records: %w", err)
		}
		a.fileEnd, a.savedPts = a.blocksEnd, 0
		return nil
	}
	frame, err := encodeBlock(sc.frame[:0], bd, a.c.schema.Scheme, a.c.db.opts.Compress)
	if err != nil {
		return err
	}
	sc.frame = frame
	if a.masked {
		maskBlock(frame, a.salt, a.blocksEnd)
	}
	var copyRec []byte
	if a.fileEnd > a.blocksEnd && saved > 0 {
		copyAt := max(a.fileEnd, a.blocksEnd+int64(len(frame))) // see overwriteRecords
		body := encodePoints(&blockData{ts: bd.ts[:saved], vals: sliceTo(bd.vals, saved), strs: sliceTo(bd.strs, saved)}, 0, a.c.schema.Scheme)
		copyRec = appendRecord(nil, recordCopy, saved, body, a.masked, a.salt, copyAt)
	}
	if err := overwriteRecords(a.c.pool, a.pendingPath, a.blocksEnd, a.fileEnd, copyRec, frame); err != nil {
		return err
	}
	if copyRec == nil && a.fileEnd > a.blocksEnd+int64(len(frame)) {
		// unreadable records behind the block: cut them off
		if err := a.c.pool.Truncate(a.pendingPath, a.blocksEnd+int64(len(frame))); err != nil {
			return fmt.Errorf("tsdb: cut behind block: %w", err)
		}
	}
	a.blocksEnd += int64(len(frame))
	a.fileEnd = a.blocksEnd
	a.blocksPts += int64(len(bd.ts))
	a.blocksMax = bd.ts[len(bd.ts)-1]
	a.savedPts = 0
	a.mem.release()
	if a.dataOff == chunkDataOff {
		next := chunkSlot{seq: a.slotSeq + 1, end: a.blocksEnd, points: a.blocksPts, max: a.blocksMax}
		if err := writeSlot(a.c.pool, a.pendingPath, next); err != nil {
			return err
		}
		a.slotSeq = next.seq
	}
	return nil
}

func sliceTo[T any](s []T, n int) []T {
	if len(s) < n {
		return s
	}
	return s[:n]
}

// finalizePending seals the open block and renames the pending chunk to its
// final <min>_<max>.tsb name, registering it in the column's chunk set.
func (a *appender) finalizePending() error {
	if a.pendingPath == "" {
		return nil
	}
	if err := a.sealBlock(); err != nil {
		return err
	}
	if a.blocksPts == 0 {
		// empty pending chunk (never happens in practice) — remove it
		_ = a.c.pool.Remove(a.pendingPath)
		a.forgetPending()
		return nil
	}
	finalPath := filepath.Join(a.c.dir, finalChunkName(a.pendingMin, a.blocksMax))
	if err := a.c.pool.Rename(a.pendingPath, finalPath); err != nil {
		return fmt.Errorf("tsdb: finalize appended chunk: %w", err)
	}
	a.c.chunks = append(a.c.chunks, chunkInfo{
		path:      finalPath,
		minMillis: a.pendingMin,
		maxMillis: a.blocksMax,
		sizeBytes: a.blocksEnd,
		dataOff:   a.dataOff,
		salt:      a.salt,
		masked:    a.masked,
	})
	a.forgetPending()
	return nil
}

// dropIfBefore removes the open pending chunk and the points still in memory, if
// all of them are older than ts. A pending chunk with a newer point stays
// untouched, because a chunk is only removed as a whole. Caller holds c.mu.
func (a *appender) dropIfBefore(ts int64) (bool, error) {
	if a.pendingPath == "" || a.newest() >= ts {
		return false, nil
	}
	if err := a.c.pool.Remove(a.pendingPath); err != nil {
		return false, fmt.Errorf("tsdb: remove pending chunk: %w", err)
	}
	a.forgetPending()
	return true, nil
}

// pendingChunk returns a chunkInfo describing the sealed blocks of the open pending chunk, if it has any, so the read
// path scans them like those of a finalized chunk. sizeBytes is where they end, so the reader needs no Fstat. The open
// block, saved or in memory, is read separately (loadOpenBlock).
func (a *appender) pendingChunk() (chunkInfo, bool) {
	if a.pendingPath == "" || a.blocksPts == 0 {
		return chunkInfo{}, false
	}
	return chunkInfo{
		path:      a.pendingPath,
		minMillis: a.pendingMin,
		maxMillis: a.blocksMax,
		pending:   true,
		sizeBytes: a.blocksEnd,
		dataOff:   a.dataOff,
		salt:      a.salt,
		masked:    a.masked,
	}, true
}

// openChunk describes the open pending chunk for the statistics: its bounds
// include the points still in memory and sizeBytes is what its file holds so
// far. ok is false when there is no pending chunk. A pending chunk always holds
// at least one point, sealed, saved or in memory.
func (a *appender) openChunk() (chunkInfo, bool) {
	if a.pendingPath == "" {
		return chunkInfo{}, false
	}
	return chunkInfo{
		path:      a.pendingPath,
		minMillis: a.pendingMin,
		maxMillis: a.newest(),
		pending:   true,
		sizeBytes: a.fileEnd,
	}, true
}
