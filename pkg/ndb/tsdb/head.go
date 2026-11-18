// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"

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
//
// The WAL is read when the head is needed first, by a write, a read or a compaction, not when the column is opened.
type head struct {
	path   string
	fs     *ndb.FilePool
	vals   *values // through which records go into a file (values.write)
	scheme Scheme
	// report reports a damaged place of the WAL (Column.reportDamaged); nil reports nothing
	report func(path string, offset int64, what string)

	mu sync.RWMutex
	// ready is set once the WAL is replayed; size is what the listing of the column found of it, so that an empty or
	// missing WAL is not opened at all
	ready atomic.Bool
	size  int64
	tree  *btree.BTreeG[headEntry]
	// closed is set by close; the WAL itself has no handle of its own, it is read and written through the shared
	// FilePool like every chunk, so that an open column holds no descriptor outside the pool
	closed bool
	// walSz is where the next record goes. cutAt, if not -1, is where a rest, which a crash tore, begins: the next
	// append cuts the WAL there first, so that no rest of it is ever read as a record or as a damaged place. legacy is
	// set for a WAL in the format before N23, which the next append or compaction rewrites first (convertLocked).
	walSz  int64
	cutAt  int64
	legacy bool
	// bytesLogged is the raw payload volume appended since the last flush to a
	// chunk; used by the compaction threshold.
	bytesLogged int64
	tombstones  int
	// seq is the number of the newest mutation (headEntry.seq)
	seq uint64
}

func lessHeadEntry(a, b headEntry) bool { return a.ts < b.ts }

// head.wal in the format of this version (N23) begins with a mark, which a version before N23 reads as an intact
// record of a kind it does not know, so that it stops with an error instead of writing over it:
//
//	| Kind     1 | walMarkKind
//	| Len      4 | 9 (LE)
//	| CRC32    4 | IEEE over the payload
//	| Payload  9 | "NAGO" "NTSW" Version (walVersion)
//
// and goes on with records:
//
//	| Marker   4 | walMarker (LE)
//	| Kind     1 | 0=put, 1=tombstone
//	| Len      4 | length of the payload below (LE)
//	| CRC32    4 | IEEE over Kind, Len and Payload
//	| Payload  n |
//
// put payload: varint ts (zig-zag) | for string scheme: WriteSlice(str); else varint val (zig-zag)
// tombstone payload: varint ts (zig-zag)
//
// The CRC covers kind and length, and a run of zero bytes is no record, because the marker is not zero. A damaged
// record or region is reported and costs only itself: the next intact record behind it is found by its marker, and
// every intact record counts in its order. A rest torn by a crash is not reported, and the next append cuts it off. An
// intact record of a kind this version does not know is errUnknownFormat, as is a mark with its CRC and a later
// version: the head does not load, and nothing of it is overwritten. nago tells this format from the one before N23
// by the magic and the format in the payload of the mark: at least 7 of their 8 bytes in place, so that one damaged
// byte of the mark costs no record; a damaged mark is reported at offset 0 and read as walVersion.
//
// A WAL without a record is empty or missing; the mark goes in front of its first record.
//
// The format before N23 (8ae1daaee) has no mark, and its records are | Kind 1 | Len 4 | CRC32 4 over the payload |
// Payload n |. Such a WAL, when it is not empty, is read up to its first damaged record, which is reported, a record of
// a kind other than put or tombstone included; nothing behind it counts. A record, whose length reaches over the end, is
// a torn rest, as in 8ae1daaee. Before the first append and before a compaction (never by an open or a read), nago
// rewrites it in the format of this version (convertLocked).
const (
	headKindPut  byte = 0
	headKindTomb byte = 1

	walMarkKind byte = 0x57 // 'W'
	walMagic         = "NAGO"
	walFormat        = "NTSW" // Nago Time Series WAL
	walVersion  byte = 1
	walMarkLen       = 9 + 9
	walMarker        = uint32(0x31C9A85E)
	walRecHdr        = 4 + 1 + 4 + 4

	headOldRecHdr = 9 // the format before N23: kind(1)+len(4)+crc(4)
)

// the text of the reports of a WAL (R-DEC-TS-NAGO-DAMAGE)
const reportWAL = "tsdb: damaged record of the head skipped"

var walMarkerBytes = binary.LittleEndian.AppendUint32(nil, walMarker)

// walMark is the mark at the beginning of a WAL.
func walMark() []byte {
	m := make([]byte, walMarkLen)
	m[0] = walMarkKind
	binary.LittleEndian.PutUint32(m[1:], 9)
	copy(m[9:], walMagic)
	copy(m[13:], walFormat)
	m[17] = walVersion
	binary.LittleEndian.PutUint32(m[5:], crc32.ChecksumIEEE(m[9:]))
	return m
}

// isWALFormat reports whether data, a non-empty WAL, is in the format of this version: its mark is there, at least 7
// of the 8 bytes of magic and format in place, or data is a mark torn by a crash.
func isWALFormat(data []byte) bool {
	if len(data) < walMarkLen {
		return bytes.HasPrefix(walMark(), data)
	}
	same := 0
	want := walMagic + walFormat
	for i := range 8 {
		if data[9+i] == want[i] {
			same++
		}
	}
	return same >= 7
}

// openHead opens the head of the column in dir and replays its WAL at once.
func openHead(dir string, pool *ndb.FilePool, scheme Scheme) (*head, error) {
	h := newHead(dir, pool, scheme, -1)
	if err := h.ensure(); err != nil {
		return nil, err
	}
	return h, nil
}

// newHead returns the head of the column in dir, whose WAL is replayed when it is needed first. size is the size of
// the WAL from the listing of the column, 0 for none, -1 for unknown.
func newHead(dir string, fs *ndb.FilePool, scheme Scheme, size int64) *head {
	return &head{
		path:   filepath.Join(dir, headWALName),
		fs:     fs,
		vals:   &values{pool: fs},
		scheme: scheme,
		size:   size,
		cutAt:  -1,
	}
}

// ensure replays the WAL, if that did not happen yet.
func (h *head) ensure() error {
	if h.ready.Load() {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.ensureLocked()
}

func (h *head) ensureLocked() error {
	if h.ready.Load() {
		return nil
	}
	h.tree = btree.NewBTreeG[headEntry](lessHeadEntry)
	h.walSz, h.cutAt, h.legacy, h.seq, h.tombstones, h.bytesLogged = 0, -1, false, 0, 0, 0
	if h.size != 0 {
		if err := h.replay(); err != nil {
			h.tree = btree.NewBTreeG[headEntry](lessHeadEntry)
			h.tombstones = 0
			return err
		}
	}
	h.ready.Store(true)
	return nil
}

func (h *head) reportAt(off int64) {
	if h.report != nil {
		h.report(h.path, off, reportWAL)
	}
}

// replay rebuilds the in-memory tree from the WAL. In the format of this version, it skips each damaged record or
// region and reports it, goes on with the next intact record, and leaves a torn rest for the next append to cut off;
// an intact record of a kind of a later version and a mark of a later version are errUnknownFormat: the head does not
// load, and nothing behind it is overwritten. A WAL in the format before N23 is read by replayOld.
func (h *head) replay() error {
	data, err := h.fs.ReadFile(h.path)
	if err != nil {
		if isNotExist(err) {
			return nil
		}
		return fmt.Errorf("tsdb: read head wal: %w", err)
	}
	if len(data) == 0 {
		return nil
	}
	if !isWALFormat(data) {
		h.legacy = true
		h.walSz = int64(h.replayOld(data))
		return nil
	}
	n := len(data)
	if n < walMarkLen {
		h.cutAt = 0 // a mark torn by a crash
		return nil
	}
	intact := data[0] == walMarkKind && binary.LittleEndian.Uint32(data[1:]) == 9 &&
		crc32.ChecksumIEEE(data[9:walMarkLen]) == binary.LittleEndian.Uint32(data[5:]) &&
		string(data[9:13]) == walMagic && string(data[13:17]) == walFormat
	if intact && data[17] > walVersion {
		return fmt.Errorf("%w: head wal %s has version %d", errUnknownFormat, h.path, data[17])
	}
	if !intact || data[17] != walVersion {
		h.reportAt(0)
	}
	off := walMarkLen
	region := -1
	for off < n {
		kind, payload, size, ok := parseWALRecord(data, off)
		if !ok {
			if region < 0 {
				region = off
			}
			if off = nextWALRecord(data, off+1); off < 0 {
				break
			}
			continue
		}
		if region >= 0 {
			h.reportAt(int64(region))
			region = -1
		}
		if kind != headKindPut && kind != headKindTomb {
			// a record of a later version: the head fails to load, so that no write goes over it
			return fmt.Errorf("%w: record of kind %d in %s at %d", errUnknownFormat, kind, h.path, off)
		}
		if err := h.applyRecord(kind, payload); err != nil {
			h.reportAt(int64(off)) // intact, but it does not decode
		}
		off += size
	}
	h.walSz = int64(n)
	if region >= 0 {
		if tornWALRecord(data[region:]) {
			h.cutAt, h.walSz = int64(region), int64(region)
		} else {
			h.reportAt(int64(region))
		}
	}
	return nil
}

// parseWALRecord returns the intact record at data[off:] in the format of this version: its kind, its payload and
// its size.
func parseWALRecord(data []byte, off int) (kind byte, payload []byte, size int, ok bool) {
	b := data[off:]
	if len(b) < walRecHdr || binary.LittleEndian.Uint32(b) != walMarker {
		return 0, nil, 0, false
	}
	l := uint64(binary.LittleEndian.Uint32(b[5:]))
	if l > uint64(len(b)-walRecHdr) {
		return 0, nil, 0, false
	}
	size = walRecHdr + int(l)
	crc := crc32.NewIEEE()
	crc.Write(b[4:9])
	crc.Write(b[walRecHdr:size])
	if crc.Sum32() != binary.LittleEndian.Uint32(b[9:]) {
		return 0, nil, 0, false
	}
	return b[4], b[walRecHdr:size], size, true
}

// nextWALRecord returns the offset of the first intact record from off on, or -1.
func nextWALRecord(data []byte, off int) int {
	for off < len(data) {
		i := bytes.Index(data[off:], walMarkerBytes)
		if i < 0 {
			return -1
		}
		off += i
		if _, _, _, ok := parseWALRecord(data, off); ok {
			return off
		}
		off++
	}
	return -1
}

// tornWALRecord reports whether b, the rest of a WAL behind its last intact record, is what a crash leaves of a record
// it appended: the beginning of one, whose length reaches over the end of the file.
func tornWALRecord(b []byte) bool {
	if len(b) < 4 {
		return bytes.HasPrefix(walMarkerBytes, b)
	}
	if binary.LittleEndian.Uint32(b) != walMarker {
		return false
	}
	return len(b) < walRecHdr || uint64(binary.LittleEndian.Uint32(b[5:])) > uint64(len(b)-walRecHdr)
}

// replayOld applies the records of a WAL in the format before N23 up to its first damaged one, which it reports, and
// returns where they end; a record, whose length reaches over the end, is a torn rest and not reported.
func (h *head) replayOld(data []byte) int {
	return walOldRecords(data, func(kind byte, payload []byte) bool {
		return h.applyRecord(kind, payload) == nil
	}, func(off int) { h.reportAt(int64(off)) })
}

// walOldRecords calls fn with every record of a WAL in the format before N23 up to the first damaged one, at which it
// calls damaged; fn returns false for a record that does not decode, which counts as damaged. It returns where the
// records end.
func walOldRecords(data []byte, fn func(kind byte, payload []byte) bool, damaged func(off int)) int {
	off := 0
	for off+headOldRecHdr <= len(data) {
		kind := data[off]
		l := uint64(binary.LittleEndian.Uint32(data[off+1:]))
		crc := binary.LittleEndian.Uint32(data[off+5:])
		start := off + headOldRecHdr
		if uint64(start)+l > uint64(len(data)) {
			break // a torn rest
		}
		payload := data[start : start+int(l)]
		if crc32.ChecksumIEEE(payload) != crc || kind != headKindPut && kind != headKindTomb || !fn(kind, payload) {
			// damaged; the kind lies outside the CRC there, so a kind it does not know is damaged too
			damaged(off)
			break
		}
		off = start + int(l)
	}
	return off
}

// convert is convertLocked under the lock of the head.
func (h *head) convert() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureLocked(); err != nil {
		return err
	}
	return h.convertLocked()
}

// convertLocked rewrites a WAL in the format before N23 in the format of this version, with its records up to the
// first damaged one, before the first append and before a compaction (R-DEC-TS-NAGO-WAL). It writes head.wal.new
// through values, syncs it with the directory, renames it to head.wal and syncs the directory, so that a crash leaves
// the old or the new WAL complete; the open of a column removes a head.wal.new left by a crash. If a step fails, the
// caller fails, the old WAL stays as it is, and the next append or compaction tries again. Caller holds h.mu and has
// replayed the WAL.
func (h *head) convertLocked() error {
	if !h.legacy {
		return nil
	}
	data, err := h.fs.ReadFile(h.path)
	if err != nil && !isNotExist(err) {
		return fmt.Errorf("tsdb: read head wal: %w", err)
	}
	var wal []byte
	walOldRecords(data, func(kind byte, payload []byte) bool {
		if _, err := h.decodeRecord(kind, payload); err != nil {
			return false
		}
		if len(wal) == 0 {
			wal = walMark()
		}
		wal = walFrame(wal, kind, payload)
		return true
	}, func(int) {})
	dir := filepath.Dir(h.path)
	newPath := h.path + newSuffix
	if err := h.vals.write(newPath, wal, -1, nil); err != nil {
		return fmt.Errorf("tsdb: rewrite head wal: %w", err)
	}
	if err := h.fs.SyncAll(newPath, dir); err != nil {
		_ = h.fs.Remove(newPath)
		return fmt.Errorf("tsdb: rewrite head wal: %w", err)
	}
	if err := h.fs.Rename(newPath, h.path); err != nil {
		_ = h.fs.Remove(newPath)
		return fmt.Errorf("tsdb: rewrite head wal: %w", err)
	}
	if err := h.fs.SyncAll(dir); err != nil {
		return fmt.Errorf("tsdb: rewrite head wal: %w", err)
	}
	h.legacy, h.cutAt = false, -1
	h.walSz, h.size = int64(len(wal)), int64(len(wal))
	return nil
}

// applyRecord applies a record of a known kind (see replay).
func (h *head) applyRecord(kind byte, payload []byte) error {
	e, err := h.decodeRecord(kind, payload)
	if err != nil {
		return err
	}
	h.setEntry(e)
	return nil
}

// decodeRecord decodes the payload of a record of a known kind.
func (h *head) decodeRecord(kind byte, payload []byte) (headEntry, error) {
	rb := xbytes.Buffer{Buf: payload}
	ts, err := readVarint(&rb)
	if err != nil {
		return headEntry{}, err
	}
	if kind == headKindTomb {
		return headEntry{ts: ts, tombstone: true}, nil
	}
	e := headEntry{ts: ts}
	if h.scheme == SchemeString {
		if e.str, err = rb.ReadString(); err != nil {
			return headEntry{}, err
		}
	} else if e.val, err = readVarint(&rb); err != nil {
		return headEntry{}, err
	}
	return e, nil
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

// walFrame appends the record of kind and payload in the format of this version to dst.
func walFrame(dst []byte, kind byte, payload []byte) []byte {
	var hdr [walRecHdr]byte
	binary.LittleEndian.PutUint32(hdr[0:], walMarker)
	hdr[4] = kind
	binary.LittleEndian.PutUint32(hdr[5:], uint32(len(payload)))
	crc := crc32.NewIEEE()
	crc.Write(hdr[4:9])
	crc.Write(payload)
	binary.LittleEndian.PutUint32(hdr[9:], crc.Sum32())
	return append(append(dst, hdr[:]...), payload...)
}

func (h *head) append(e headEntry) error {
	kind, payload := h.record(e)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return errColumnGone
	}
	if err := h.ensureLocked(); err != nil {
		return err
	}
	if err := h.convertLocked(); err != nil {
		return err
	}
	if h.cutAt >= 0 {
		if err := h.fs.Truncate(h.path, h.cutAt); err != nil {
			return fmt.Errorf("tsdb: append head wal: %w", err)
		}
		h.walSz, h.cutAt = h.cutAt, -1
	}
	var rec []byte
	if h.walSz == 0 {
		rec = walMark()
	}
	rec = walFrame(rec, kind, payload)
	if err := h.vals.write(h.path, rec, h.walSz, nil); err != nil {
		return fmt.Errorf("tsdb: append head wal: %w", err)
	}
	h.walSz += int64(len(rec))
	h.size = h.walSz
	h.bytesLogged += int64(len(payload))
	h.setEntry(e)
	return nil
}

// snapshot returns a stable ordered copy of the current head entries in
// [min,max] (inclusive). Used by reads and by compaction. Cheap relative to
// chunk data because the head is bounded by the flush threshold.
func (h *head) snapshot(min, max int64) []headEntry {
	_ = h.ensure() // an unreadable WAL is an empty head here; the callers, which can report it, call ensure first
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
	_ = h.ensure() // an unreadable WAL is an empty head here; the callers, which can report it, call ensure first
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
	_ = h.ensure() // an unreadable WAL is an empty head here; the callers, which can report it, call ensure first
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.tree.Get(headEntry{ts: ts})
}

// maxTS returns the largest timestamp held in the head, if any.
func (h *head) maxTS() (int64, bool) {
	_ = h.ensure() // an unreadable WAL is an empty head here; the callers, which can report it, call ensure first
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
	_ = h.ensure() // an unreadable WAL is an empty head here; the callers, which can report it, call ensure first
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.tree.Len() == 0 {
		return 0, false
	}
	e, ok := h.tree.Min()
	return e.ts, ok
}

func (h *head) len() int {
	_ = h.ensure() // an unreadable WAL is an empty head here; the callers, which can report it, call ensure first
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.tree.Len()
}

func (h *head) tombstoneCount() int {
	_ = h.ensure() // an unreadable WAL is an empty head here; the callers, which can report it, call ensure first
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
// stays in the tree and in the new WAL, in the order of
// its sequence numbers, so that replay rebuilds the same tree. The new WAL replaces the old one by a rename after it
// is on the disk, so that a crash leaves one of both (see foldBegin).
func (h *head) resetFolded(mark uint64) error {
	dir := filepath.Dir(h.path)
	f, err := h.foldBegin(mark, filepath.Join(dir, headWALName+".fold"+newSuffix), nil)
	if err != nil {
		return err
	}
	if err := h.fs.SyncAll(f.newPath, dir); err != nil {
		f.abort()
		return fmt.Errorf("tsdb: sync new head wal: %w", err)
	}
	if err := f.finish(); err != nil {
		return err
	}
	if err := h.fs.SyncAll(dir); err != nil {
		return fmt.Errorf("tsdb: sync column dir: %w", err)
	}
	return nil
}

// headFold is a fold of the head in progress: the new WAL holds the entries after mark as they were at foldBegin.
// Writers go on meanwhile, into the old WAL; finish moves what they wrote behind the new WAL before it replaces the
// old one.
type headFold struct {
	h       *head
	mark    uint64
	oldSize int64 // the size of the valid part of the old WAL at foldBegin
	newPath string
	newSize int64
}

// foldBegin writes the entries after mark, in the order of their sequence numbers, as a new WAL at newPath, in the
// format of this version: a WAL in the format before N23 is rewritten first (convertLocked), so that what writers put
// into the old WAL meanwhile has the same format as the new one. The caller syncs it before it replaces the old one.
// It holds the lock of the head only to take them. k is the compaction, whose file newPath is, or nil.
func (h *head) foldBegin(mark uint64, newPath string, k *compaction) (*headFold, error) {
	h.mu.Lock()
	if err := h.ensureLocked(); err != nil {
		h.mu.Unlock()
		return nil, err
	}
	if err := h.convertLocked(); err != nil {
		h.mu.Unlock()
		return nil, err
	}
	var kept []headEntry
	h.tree.Scan(func(e headEntry) bool {
		if e.seq > mark {
			kept = append(kept, e)
		}
		return true
	})
	f := &headFold{h: h, mark: mark, oldSize: h.walSz, newPath: newPath}
	h.mu.Unlock()

	slices.SortFunc(kept, func(a, b headEntry) int { return cmp.Compare(a.seq, b.seq) })
	var wal []byte
	for _, e := range kept {
		if len(wal) == 0 {
			wal = walMark()
		}
		kind, payload := h.record(e)
		wal = walFrame(wal, kind, payload)
	}
	f.newSize = int64(len(wal))
	if err := h.vals.write(newPath, wal, -1, k); err != nil {
		f.abort()
		return nil, fmt.Errorf("tsdb: write new head wal: %w", err)
	}
	return f, nil
}

// finish appends what was written to the old WAL since foldBegin to the new one, renames the new one over the old
// one and removes the entries up to mark from the tree. It holds the lock of the head, so that no writer comes in
// between. The appended records were not synced in the old WAL either.
func (f *headFold) finish() error {
	h := f.h
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cutAt >= 0 {
		// no writer came since foldBegin, which would have cut the torn rest off: there is nothing to move
		h.walSz, h.cutAt = f.oldSize, -1
	}
	added, err := appendWALTail(h.vals, h.path, f.oldSize, h.walSz, f.newPath, f.newSize)
	if err != nil {
		return err
	}
	f.newSize += added
	if err := h.fs.Rename(f.newPath, h.path); err != nil {
		return fmt.Errorf("tsdb: replace head wal: %w", err)
	}
	tree := btree.NewBTreeG[headEntry](lessHeadEntry)
	var logged int64
	tombstones := 0
	h.tree.Scan(func(e headEntry) bool {
		if e.seq > f.mark {
			tree.Set(e)
			_, payload := h.record(e)
			logged += int64(len(payload))
			if e.tombstone {
				tombstones++
			}
		}
		return true
	})
	h.tree = tree
	h.walSz = f.newSize
	h.size = f.newSize
	h.bytesLogged = logged
	h.tombstones = tombstones
	return nil
}

// abort removes the new WAL; the head stays as it was.
func (f *headFold) abort() { _ = f.h.fs.Remove(f.newPath) }

// appendWALTail appends the bytes from..to of the WAL at path to the WAL at newPath, which has newSize bytes, and
// returns how many bytes it appended. Into an empty new WAL in the format of this version goes the mark first, when
// the old one is in that format, so that the records behind it are read in their format.
func appendWALTail(vals *values, path string, from, to int64, newPath string, newSize int64) (int64, error) {
	if to <= from {
		return 0, nil
	}
	tail := make([]byte, to-from)
	if _, err := vals.pool.ReadAt(path, tail, from); err != nil {
		return 0, fmt.Errorf("tsdb: read head wal: %w", err)
	}
	if newSize == 0 && from > 0 {
		var mark [walMarkLen]byte
		n, _ := vals.pool.ReadAt(path, mark[:], 0)
		if n == walMarkLen && isWALFormat(mark[:]) {
			tail = append(walMark(), tail...)
		}
	}
	if err := vals.write(newPath, tail, newSize, nil); err != nil {
		return 0, fmt.Errorf("tsdb: append to new head wal: %w", err)
	}
	return int64(len(tail)), nil
}

func (h *head) close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.closed {
		h.closed = true
		h.fs.Evict(h.path)
	}
	return nil
}
