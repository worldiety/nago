// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"cmp"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/tidwall/btree"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/xbytes"
)

const headWALName = "head.wal"

// headEntry is one logical mutation for a single timestamp held in memory. A
// tombstone marks a deleted timestamp so that a value present in a sealed chunk
// is masked on read until compaction physically removes it.
type headEntry struct {
	ts        int64
	val       int64  // numeric value or enum id (as int64); ignored for tombstones
	str       string // string scheme value
	tombstone bool
	// seq orders the mutations of the head: every put or tombstone gets the next number, also in replay. It lives only
	// in memory, so that a fold removes exactly the entries of its snapshot and keeps every later one (resetFolded).
	seq uint64
}

// head is the mutable write layer of a column: an append-only WAL plus an
// in-memory btree keyed by timestamp (newest-wins). Reads merge the head over
// the sealed chunks. Its own lock guards the tree and the WAL. Writers of
// corrections and tombstones do not hold the lock of the column, so they may
// write while a compaction folds a snapshot of the head; the compaction
// therefore removes only the entries of its snapshot (resetFolded).
type head struct {
	path   string
	pool   *ndb.FilePool
	scheme Scheme

	mu   sync.RWMutex
	tree *btree.BTreeG[headEntry]
	// closed is set by close; the WAL itself has no handle of its own, it is read and written through the shared
	// FilePool like every chunk, so that an open column holds no descriptor outside the pool
	closed bool
	walSz  int64
	// bytesLogged is the raw payload volume appended since the last flush to a
	// chunk; used by the compaction threshold.
	bytesLogged int64
	tombstones  int
	// seq is the number of the newest mutation (headEntry.seq)
	seq uint64
}

func lessHeadEntry(a, b headEntry) bool { return a.ts < b.ts }

// head WAL record layout (per record):
//
//	| Kind  1 | 0=put, 1=tombstone
//	| Len   4 | length of the payload below (LE)
//	| CRC   4 | IEEE over payload
//	| Payload n |
//
// put payload: varint ts (zig-zag) | for string scheme: WriteSlice(str); else varint val (zig-zag)
// tombstone payload: varint ts (zig-zag)
const (
	headKindPut  byte = 0
	headKindTomb byte = 1
	headRecHdr        = 9 // kind(1)+len(4)+crc(4)
)

func openHead(dir string, pool *ndb.FilePool, scheme Scheme) (*head, error) {
	h := &head{
		path:   filepath.Join(dir, headWALName),
		pool:   pool,
		scheme: scheme,
		tree:   btree.NewBTreeG[headEntry](lessHeadEntry),
	}
	if err := h.replay(); err != nil {
		return nil, err
	}
	if _, err := os.Stat(h.path); os.IsNotExist(err) {
		if err := os.WriteFile(h.path, nil, 0644); err != nil {
			return nil, fmt.Errorf("tsdb: create head wal: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("tsdb: stat head wal: %w", err)
	}
	return h, nil
}

// replay rebuilds the in-memory tree from the WAL, stopping at the first
// corrupt/truncated record (house rule: continue operating, drop torn tail).
func (h *head) replay() error {
	data, err := os.ReadFile(h.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("tsdb: read head wal: %w", err)
	}
	off := 0
	for off+headRecHdr <= len(data) {
		kind := data[off]
		l := int(binary.LittleEndian.Uint32(data[off+1:]))
		crc := binary.LittleEndian.Uint32(data[off+5:])
		start := off + headRecHdr
		if l < 0 || start+l > len(data) {
			break
		}
		payload := data[start : start+l]
		if crc32.ChecksumIEEE(payload) != crc {
			break
		}
		if err := h.applyRecord(kind, payload); err != nil {
			break
		}
		off = start + l
	}
	h.walSz = int64(off)
	return nil
}

func (h *head) applyRecord(kind byte, payload []byte) error {
	rb := xbytes.Buffer{Buf: payload}
	ts, err := readVarint(&rb)
	if err != nil {
		return err
	}
	switch kind {
	case headKindTomb:
		h.setEntry(headEntry{ts: ts, tombstone: true})
	case headKindPut:
		var e headEntry
		e.ts = ts
		if h.scheme == SchemeString {
			s, err := rb.ReadString()
			if err != nil {
				return err
			}
			e.str = s
		} else {
			v, err := readVarint(&rb)
			if err != nil {
				return err
			}
			e.val = v
		}
		h.setEntry(e)
	default:
		return errCorruptBlock
	}
	return nil
}

// setEntry inserts or replaces an entry in the tree, maintaining the tombstone
// counter, and gives it the next sequence number. Caller holds the appropriate
// lock (or is in replay, single-threaded).
func (h *head) setEntry(e headEntry) {
	h.seq++
	e.seq = h.seq
	prev, replaced := h.tree.Set(e)
	if replaced {
		if prev.tombstone && !e.tombstone {
			h.tombstones--
		} else if !prev.tombstone && e.tombstone {
			h.tombstones++
		}
	} else if e.tombstone {
		h.tombstones++
	}
}

// put appends a put record and updates the tree. For string scheme str is used;
// otherwise val (numeric value or enum id).
func (h *head) put(ts int64, val int64, str string) error {
	return h.append(headEntry{ts: ts, val: val, str: str})
}

// tombstone appends a tombstone record and updates the tree.
func (h *head) tombstone(ts int64) error {
	return h.append(headEntry{ts: ts, tombstone: true})
}

// record is the WAL record of e: its kind and its payload.
func (h *head) record(e headEntry) (byte, []byte) {
	pb := xbytes.Buffer{}
	writeVarint(&pb, e.ts)
	if e.tombstone {
		return headKindTomb, pb.Buf[:pb.Pos]
	}
	if h.scheme == SchemeString {
		_, _ = pb.WriteString(e.str)
	} else {
		writeVarint(&pb, e.val)
	}
	return headKindPut, pb.Buf[:pb.Pos]
}

// frame appends the framed WAL record of kind and payload to dst.
func frame(dst []byte, kind byte, payload []byte) []byte {
	var hdr [headRecHdr]byte
	hdr[0] = kind
	binary.LittleEndian.PutUint32(hdr[1:], uint32(len(payload)))
	binary.LittleEndian.PutUint32(hdr[5:], crc32.ChecksumIEEE(payload))
	return append(append(dst, hdr[:]...), payload...)
}

func (h *head) append(e headEntry) error {
	kind, payload := h.record(e)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return errColumnGone
	}
	rec := frame(nil, kind, payload)
	if err := h.writeAt(rec, h.walSz); err != nil {
		return fmt.Errorf("tsdb: append head wal: %w", err)
	}
	h.walSz += int64(len(rec))
	h.bytesLogged += int64(len(payload))
	h.setEntry(e)
	return nil
}

// snapshot returns a stable ordered copy of the current head entries in
// [min,max] (inclusive). Used by reads and by compaction. Cheap relative to
// chunk data because the head is bounded by the flush threshold.
func (h *head) snapshot(min, max int64) []headEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var out []headEntry
	h.tree.Ascend(headEntry{ts: min}, func(e headEntry) bool {
		if e.ts > max {
			return false
		}
		out = append(out, e)
		return true
	})
	return out
}

// hasRange reports whether the head holds any entry (put or tombstone) with
// min <= ts <= max. It is a cheap check used to select the allocation-free read
// fast path when the head does not overlap the query range.
func (h *head) hasRange(min, max int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.tree.Len() == 0 {
		return false
	}
	found := false
	h.tree.Ascend(headEntry{ts: min}, func(e headEntry) bool {
		if e.ts > max {
			return false
		}
		found = true
		return false
	})
	return found
}

// lookup returns the head entry for exactly ts, if any.
func (h *head) lookup(ts int64) (headEntry, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.tree.Get(headEntry{ts: ts})
}

// maxTS returns the largest timestamp held in the head, if any.
func (h *head) maxTS() (int64, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.tree.Len() == 0 {
		return 0, false
	}
	e, ok := h.tree.Max()
	return e.ts, ok
}

// minTS returns the smallest timestamp held in the head, if any.
func (h *head) minTS() (int64, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.tree.Len() == 0 {
		return 0, false
	}
	e, ok := h.tree.Min()
	return e.ts, ok
}

func (h *head) len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.tree.Len()
}

func (h *head) tombstoneCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.tombstones
}

// foldMark is the sequence number of the newest entry of a snapshot: a fold of
// that snapshot removes exactly the entries up to it (resetFolded).
func foldMark(entries []headEntry) uint64 {
	var m uint64
	for _, e := range entries {
		m = max(m, e.seq)
	}
	return m
}

// resetFolded removes the entries up to mark from the tree and the WAL after a
// compaction has folded them into sealed chunks. A put or tombstone with a
// larger sequence number came after the snapshot of that compaction, because
// writers do not hold the lock of the column while they write the head; it
// stays in the tree and is written anew into the truncated WAL, in the order of
// its sequence numbers, so that replay rebuilds the same tree. Without such an
// entry the WAL is just truncated.
func (h *head) resetFolded(mark uint64) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	var kept []headEntry
	h.tree.Scan(func(e headEntry) bool {
		if e.seq > mark {
			kept = append(kept, e)
		}
		return true
	})
	slices.SortFunc(kept, func(a, b headEntry) int { return cmp.Compare(a.seq, b.seq) })
	var wal []byte
	var logged int64
	tree := btree.NewBTreeG[headEntry](lessHeadEntry)
	tombstones := 0
	for _, e := range kept {
		kind, payload := h.record(e)
		wal = frame(wal, kind, payload)
		logged += int64(len(payload))
		tree.Set(e)
		if e.tombstone {
			tombstones++
		}
	}
	if err := os.Truncate(h.path, 0); err != nil {
		return fmt.Errorf("tsdb: truncate head wal: %w", err)
	}
	if len(wal) > 0 {
		if err := h.writeAt(wal, 0); err != nil {
			return fmt.Errorf("tsdb: rewrite head wal: %w", err)
		}
	}
	h.walSz = int64(len(wal))
	h.bytesLogged = logged
	h.tombstones = tombstones
	h.tree = tree
	return nil
}

func (h *head) close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.closed {
		h.closed = true
		h.pool.Evict(h.path)
	}
	return nil
}

// writeAt writes buf at off of the WAL through the pool, which keeps the handle open for the write.
func (h *head) writeAt(buf []byte, off int64) error {
	_, err := h.pool.WriteAt(h.path, buf, off)
	return err
}
