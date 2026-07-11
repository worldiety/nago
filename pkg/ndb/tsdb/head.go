// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
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
}

// head is the mutable write layer of a column: an append-only WAL plus an
// in-memory btree keyed by timestamp (newest-wins). Reads merge the head over
// the sealed chunks. Not safe for concurrent use on its own; the owning column
// serializes access under its write lock, and reads take a read lock.
type head struct {
	path   string
	pool   *ndb.FilePool
	scheme Scheme

	mu    sync.RWMutex
	tree  *btree.BTreeG[headEntry]
	wal   *os.File
	walSz int64
	// bytesLogged is the raw payload volume appended since the last flush to a
	// chunk; used by the compaction threshold.
	bytesLogged int64
	tombstones  int
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
	f, err := os.OpenFile(h.path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("tsdb: open head wal: %w", err)
	}
	if _, err := f.Seek(h.walSz, 0); err != nil {
		f.Close()
		return nil, fmt.Errorf("tsdb: seek head wal: %w", err)
	}
	h.wal = f
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
// counter. Caller holds the appropriate lock (or is in replay, single-threaded).
func (h *head) setEntry(e headEntry) {
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
	pb := xbytes.Buffer{}
	writeVarint(&pb, ts)
	if h.scheme == SchemeString {
		_, _ = pb.WriteString(str)
	} else {
		writeVarint(&pb, val)
	}
	return h.append(headKindPut, pb.Buf[:pb.Pos], headEntry{ts: ts, val: val, str: str})
}

// tombstone appends a tombstone record and updates the tree.
func (h *head) tombstone(ts int64) error {
	pb := xbytes.Buffer{}
	writeVarint(&pb, ts)
	return h.append(headKindTomb, pb.Buf[:pb.Pos], headEntry{ts: ts, tombstone: true})
}

func (h *head) append(kind byte, payload []byte, e headEntry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	rec := make([]byte, headRecHdr+len(payload))
	rec[0] = kind
	binary.LittleEndian.PutUint32(rec[1:], uint32(len(payload)))
	binary.LittleEndian.PutUint32(rec[5:], crc32.ChecksumIEEE(payload))
	copy(rec[headRecHdr:], payload)
	if _, err := h.wal.WriteAt(rec, h.walSz); err != nil {
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

// reset truncates the WAL and clears the in-memory tree after its contents have
// been folded into sealed chunks by compaction. Caller must ensure no
// concurrent writes.
func (h *head) reset() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.wal.Truncate(0); err != nil {
		return fmt.Errorf("tsdb: truncate head wal: %w", err)
	}
	h.walSz = 0
	h.bytesLogged = 0
	h.tombstones = 0
	h.tree = btree.NewBTreeG[headEntry](lessHeadEntry)
	return nil
}

func (h *head) close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.wal != nil {
		err := h.wal.Close()
		h.wal = nil
		return err
	}
	return nil
}
