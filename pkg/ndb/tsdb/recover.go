// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"bytes"
	"fmt"
)

// openState is the state of an open chunk: where its sealed blocks end, and the saved points of its open block in the
// continuation records behind them.
type openState struct {
	path      string
	minMillis int64 // from its name
	dataOff   int64
	salt      uint64 // the key of the masks of a version 2 file
	masked    bool
	slotSeq   uint32 // the newest slot of a version 2 header, 0 for none
	blocksEnd int64
	blocksPts int64
	blocksMax int64 // valid if blocksPts > 0
	fileEnd   int64
	savedPts  int64
	savedMin  int64
	savedMax  int64
}

// points is the number of points the open chunk holds on disk.
func (s *openState) points() int64 { return s.blocksPts + s.savedPts }

// newest is its newest time on disk.
func (s *openState) newest() int64 {
	if s.savedPts > 0 {
		return s.savedMax
	}
	return s.blocksMax
}

// foundRecord is a continuation record found by recoverOpenChunk.
type foundRecord struct {
	at, end int64
	kind    byte
	pts     blockData
}

func (r *foundRecord) first() int64 { return r.pts.ts[0] }
func (r *foundRecord) last() int64  { return r.pts.ts[len(r.pts.ts)-1] }

// recovered is the outcome of recoverOpenChunk.
type recovered int

const (
	recoveredPoints  recovered = iota // the chunk holds points
	recoveredEmpty                    // the chunk holds no intact point: the caller removes it
	recoveredNothing                  // the file is not a chunk: the caller leaves it on disk and skips it
)

// recoverOpenChunk reads the open chunk at path left by an earlier life, ended by Close or by a crash, so that the
// column continues it. It reads the header and, from where its newer slot says the sealed blocks end, the rest of the
// file; a chunk of version 1 or without an intact slot is read from its first block on. Only intact frames count
// (marker, length and CRC); an intact block or record of a later version is errUnknownFormat, and the file stays as it
// is. The saved points of the open block are those of the records behind the last intact block:
//
//   - A record, whose points are not newer than that block, belongs to it: a crash came after the block was written
//     over its records and before the file was cut behind it.
//   - The newest copy record replaces the records before it: a crash came while the block was written over them.
//   - The records after it continue it in the order of the file, each newer than the one before.
//
// Then it cleans up, so that the file holds exactly that: a torn rest is cut off, and a file that holds more than its
// records (a copy behind the remains of others) gets its open block written anew as one record, in the same order as a
// seal writes a block over records (see overwriteRecords).
func recoverOpenChunk(c *Column, path string, minMillis int64, r *frameReader) (st openState, res recovered, err error) {
	fs, scheme := c.pool, c.schema.Scheme
	st = openState{path: path, minMillis: minMillis}
	size, err := statSize(fs, path)
	if err != nil {
		return st, recoveredNothing, fmt.Errorf("tsdb: stat open chunk: %w", err)
	}
	head, err := readChunkHeader(fs, path)
	dataOff, slot, haveSlot := head.dataOff, head.slot, head.haveSlot
	st.salt, st.masked = head.salt, head.masked
	if err != nil {
		if err == errCorruptChunk {
			if size < chunkDataOff {
				return st, recoveredEmpty, nil // the header of a new chunk, torn by a crash
			}
			c.reportDamaged(path, 0, reportBlock) // not a chunk: left on disk and skipped with every block in it
			return st, recoveredNothing, nil
		}
		return st, recoveredNothing, err // also a chunk of a later version, which is left as it is
	}
	st.dataOff = dataOff
	start := dataOff
	if haveSlot {
		st.slotSeq = slot.seq
		if slot.end >= dataOff && slot.end <= size && slot.points >= 0 {
			start = slot.end
			st.blocksPts = slot.points
			st.blocksMax = slot.max
		}
	}
	st.blocksEnd = start

	var recs []foundRecord
	// frames are the intact frames behind the last block, as [at, end) pairs: what lies between them is no frame
	var frames [][2]int64
	r.reset(c, path, start, size, true)
	for {
		kind, at, frame, err := r.next(frameBlock | frameRecord)
		if err != nil {
			return st, recoveredNothing, err
		}
		if kind == frameNone {
			break
		}
		if kind == frameBlock {
			h, _ := parseBlockHeader(frame)
			if h.flags&^flagsKnown != 0 {
				// an intact block of a later version: nothing is cut, and the slot is not moved over it
				return st, recoveredNothing, fmt.Errorf("%w: block flags 0x%02x in %s at %d", errUnknownFormat, h.flags, path, at)
			}
			st.blocksPts += int64(h.count)
			st.blocksMax = h.maxMillis
			st.blocksEnd = at + int64(h.frameLen)
			recs = recs[:0] // a record before a block is older than it
			frames = frames[:0]
			continue
		}
		frames = append(frames, [2]int64{at, at + int64(len(frame))})
		if unknownRecord(frame, len(frame)) {
			// a record of a later version: nothing is cut or rewritten
			return st, recoveredNothing, fmt.Errorf("%w: record of kind 0x%02x in %s at %d", errUnknownFormat, frame[recHeadLen], path, at)
		}
		rk, count, body, _ := parseRecord(frame, len(frame), st.salt, at)
		fr := foundRecord{at: at, end: at + int64(len(frame)), kind: rk}
		if count == 0 {
			continue
		}
		if !decodePoints(body, count, scheme, &fr.pts) {
			c.reportDamaged(path, at, reportRecord)
			continue
		}
		recs = append(recs, fr)
	}

	// the records of the open block
	var live []foundRecord
	for _, fr := range recs {
		if st.blocksPts > 0 && fr.first() <= st.blocksMax {
			continue // it belongs to the sealed block before it
		}
		live = append(live, fr)
	}
	from := 0
	for i, fr := range live {
		if fr.kind == recordCopy {
			from = i
		}
	}
	var chosen []foundRecord
	for _, fr := range live[from:] {
		if len(chosen) > 0 && fr.first() <= chosen[len(chosen)-1].last() {
			break // out of order: what follows is no continuation of the block
		}
		chosen = append(chosen, fr)
	}

	clean := true
	expect := st.blocksEnd
	for i, fr := range chosen {
		if fr.at != expect && (i > 0 || fr.kind == recordDelta && !framesIn(frames, expect, fr.at)) {
			// a Sync appends each record behind the one before, the first behind the blocks, and no crash leaves bytes
			// between them: what lies there is a damaged record, which the rewrite below drops. Only a seal or this
			// rewrite, which a crash stopped, leaves records or other bytes in front of the copy they wrote.
			c.reportDamaged(path, expect, reportRecord)
		}
		clean = clean && fr.at == expect
		expect = fr.end
	}
	if expect < size && !framesIn(frames, expect, size) && !crashRest(r, path, expect, size) {
		// behind the last record lies neither what a crash leaves nor an intact frame: a damaged record, which the cut
		// below drops
		c.reportDamaged(path, expect, reportRecord)
	}
	switch {
	case len(chosen) == 0:
		if size > st.blocksEnd {
			if err := fs.Truncate(path, st.blocksEnd); err != nil {
				return st, recoveredNothing, fmt.Errorf("tsdb: cut open chunk: %w", err)
			}
		}
		st.fileEnd = st.blocksEnd
	case clean:
		if size > expect {
			if err := fs.Truncate(path, expect); err != nil {
				return st, recoveredNothing, fmt.Errorf("tsdb: cut open chunk: %w", err)
			}
		}
		st.fileEnd = expect
	default:
		var all blockData
		for _, fr := range chosen {
			all.ts = append(all.ts, fr.pts.ts...)
			all.vals = append(all.vals, fr.pts.vals...)
			all.strs = append(all.strs, fr.pts.strs...)
		}
		body := encodePoints(&all, 0, scheme)
		rec := appendRecord(nil, recordCopy, len(all.ts), body, st.masked, st.salt, st.blocksEnd)
		copyRec := appendRecord(nil, recordCopy, len(all.ts), body, st.masked, st.salt, max(size, st.blocksEnd+int64(len(rec))))
		if err := overwriteRecords(&c.vals, path, st.blocksEnd, size, copyRec, rec); err != nil {
			return st, recoveredNothing, err
		}
		st.fileEnd = st.blocksEnd + int64(len(rec))
	}
	for _, fr := range chosen {
		if st.savedPts == 0 {
			st.savedMin = fr.first()
		}
		st.savedPts += int64(len(fr.pts.ts))
		st.savedMax = fr.last()
	}

	// a slot behind the blocks found, so that the next open reads less
	if dataOff == chunkDataOff && (!haveSlot || st.blocksEnd != slot.end) && st.blocksPts > 0 {
		next := chunkSlot{seq: st.slotSeq + 1, end: st.blocksEnd, points: st.blocksPts, max: st.blocksMax}
		if err := writeSlot(fs, path, next); err != nil {
			return st, recoveredNothing, err
		}
		st.slotSeq = next.seq
	}
	if st.points() == 0 {
		return st, recoveredEmpty, nil
	}
	return st, recoveredPoints, nil
}

// framesIn reports whether an intact frame lies in [from, to).
func framesIn(frames [][2]int64, from, to int64) bool {
	for _, f := range frames {
		if f[0] >= from && f[1] <= to {
			return true
		}
	}
	return false
}

// crashRest reports whether the bytes from..to at the end of the open chunk at path are what a crash leaves behind
// its last frame: zeros, which a copy record written behind the end of the file leaves in front of it (see
// overwriteRecords), then perhaps the beginning of a frame, whose length reaches over the end.
func crashRest(r *frameReader, path string, from, to int64) bool {
	rest := make([]byte, to-from)
	if n, err := r.fs.ReadAt(path, rest, from); err != nil && n < len(rest) {
		return false
	}
	i := 0
	for i < len(rest) && rest[i] == 0 {
		i++
	}
	rest = rest[i:]
	if len(rest) == 0 {
		return true
	}
	if len(rest) < len(recordMarkerBytes) {
		return bytes.HasPrefix(recordMarkerBytes, rest) || bytes.HasPrefix(blockMarkerBytes, rest)
	}
	k, frameLen := peekFrame(rest)
	if k == frameNone {
		return len(rest) < recHeadLen && (bytes.HasPrefix(rest, recordMarkerBytes) || bytes.HasPrefix(rest, blockMarkerBytes))
	}
	return frameLen > len(rest)
}

// overwriteRecords replaces the records of an open block from at to end of the file at path by frame (a block or a
// record), such that a crash of the process at any byte leaves either the records or frame readable: first copy, a
// copy record of the saved points of the block, goes behind both, to max(end, at+len(frame)), for which it is masked,
// then frame is written at at, then the file is cut behind frame. Without a copy (nil) frame is just written at at.
// Both go through vals (values.write).
func overwriteRecords(vals *values, path string, at, end int64, copyRec, frame []byte) error {
	if len(copyRec) > 0 {
		if err := vals.write(path, copyRec, max(end, at+int64(len(frame))), nil); err != nil {
			return fmt.Errorf("tsdb: write copy record: %w", err)
		}
	}
	if err := vals.write(path, frame, at, nil); err != nil {
		return fmt.Errorf("tsdb: write over records: %w", err)
	}
	if len(copyRec) > 0 {
		if err := vals.pool.Truncate(path, at+int64(len(frame))); err != nil {
			return fmt.Errorf("tsdb: cut behind records: %w", err)
		}
	}
	return nil
}
