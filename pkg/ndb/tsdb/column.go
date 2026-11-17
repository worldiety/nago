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
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

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
	// dropped is set under mu, after which this handle fails: its files are gone or belong to a new column of that name.
	dropped bool

	// app is the monotonic-append fast path (bounded memory, no rewrite).
	app appender
	// curMax is the largest timestamp known to the column across sealed chunks
	// and the append buffer. A write with ts > curMax takes the append fast
	// path; ts <= curMax is an out-of-order correction handled by the head.
	curMax     int64
	haveCurMax bool
	// headInMax is set once the newest time of the head is part of curMax: the head is read by the first write, not by
	// the open of the column.
	headInMax bool

	// lastWrite is the time of the last write by Options.Now, in unix nanoseconds, kept only with Options.IdleAfter.
	lastWrite atomic.Int64
}

// readScratch is the reusable buffer set of a read. It comes from a pool shared by all columns and goes back when the
// read ends, so that a column holds no read buffer between reads and a scan of billions of points still performs no
// per-block or per-point heap allocation.
type readScratch struct {
	bd     blockData // decode target of a block
	ob     blockData // the points of the open block
	out    blockData // merged output of the slow path
	rd     frameReader
	frame  []byte
	chunks []*chunkInfo
	pc     chunkInfo
}

var scratchPool = sync.Pool{New: func() any { return new(readScratch) }}

const maxPooledBuffer = 4 << 20 // a larger read buffer is not kept for the next read

func getScratch() *readScratch { return scratchPool.Get().(*readScratch) }

func putScratch(sc *readScratch) {
	for _, bd := range []*blockData{&sc.bd, &sc.ob, &sc.out} {
		clear(bd.strs[:cap(bd.strs)])
		bd.ts, bd.vals, bd.strs = bd.ts[:0], bd.vals[:0], bd.strs[:0]
	}
	if cap(sc.rd.buf) > maxPooledBuffer {
		sc.rd.buf = nil
	}
	sc.rd.col, sc.rd.fs, sc.rd.path = nil, nil, ""
	clear(sc.chunks[:cap(sc.chunks)])
	sc.chunks = sc.chunks[:0]
	sc.pc = chunkInfo{}
	scratchPool.Put(sc)
}

// openColumn opens the column in dir. It lists the directory, finishes a committed compaction, removes the files a
// crash left behind, takes the chunks from their names and continues the open chunk, which it reads from where its
// header says the sealed blocks end. Other chunks, the dictionary and the head are read when a read or a write needs
// them.
func openColumn(db *DB, bucket, name, dir string, schema Schema) (*Column, error) {
	c := &Column{
		db:     db,
		bucket: bucket,
		name:   name,
		dir:    dir,
		schema: schema,
		pool:   db.opts.FilePool,
	}
	c.app.c = c
	entries, err := listDir(dir)
	if err != nil {
		return nil, err
	}
	if slices.ContainsFunc(entries, func(e os.DirEntry) bool { return e.Name() == compactLogName }) {
		if err := c.redoCompaction(); err != nil {
			return nil, err
		}
		if entries, err = listDir(dir); err != nil {
			return nil, err
		}
	}
	var pending []chunkInfo
	headSize, dictSize := int64(0), int64(0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		fname := e.Name()
		path := filepath.Join(dir, fname)
		switch {
		case strings.HasSuffix(fname, newSuffix) || strings.HasPrefix(fname, ".tmp-"):
			// a compaction or a schema, which was not committed
			if err := c.pool.Remove(path); err != nil {
				return nil, fmt.Errorf("tsdb: remove rest: %w", err)
			}
			continue
		case fname == headWALName:
			headSize = entrySize(e)
		case fname == enumDictName:
			dictSize = entrySize(e)
		}
		info, ok := parseChunkName(fname)
		if !ok {
			continue
		}
		info.path = path
		if info.pending {
			pending = append(pending, info)
			continue
		}
		info.sizeBytes = max(0, entrySize(e))
		c.chunks = append(c.chunks, info)
	}
	c.head = newHead(dir, c.pool, schema.Scheme, headSize)
	if schema.Scheme == SchemeEnum {
		c.dict = newEnumDict(dir, c.pool, dictSize)
	}
	if err := c.resumePending(pending); err != nil {
		return nil, err
	}

	// curMax from the finalized chunks and the continued open chunk; the head adds to it at the first write
	for _, ci := range c.chunks {
		if !c.haveCurMax || ci.maxMillis > c.curMax {
			c.curMax = ci.maxMillis
			c.haveCurMax = true
		}
	}
	if c.app.pendingPath != "" && (!c.haveCurMax || c.app.newest() > c.curMax) {
		c.curMax = c.app.newest()
		c.haveCurMax = true
	}
	return c, nil
}

func listDir(dir string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("tsdb: list column: %w", err)
	}
	return entries, nil
}

// entrySize is the size of a listed file, -1 if it is not known.
func entrySize(e os.DirEntry) int64 {
	fi, err := e.Info()
	if err != nil {
		return -1
	}
	return fi.Size()
}

func sortChunks(chunks []chunkInfo) {
	sort.Slice(chunks, func(i, j int) bool {
		if chunks[i].minMillis != chunks[j].minMillis {
			return chunks[i].minMillis < chunks[j].minMillis
		}
		return chunks[i].maxMillis < chunks[j].maxMillis
	})
}

// resumePending lets the appender continue the pending chunk, which an earlier
// process left open by DB.Close or a crash, so that a restart does not begin a
// new chunk file: its intact blocks and records stay, a torn rest is cut off, and the
// split policy ends it as if the process had never stopped. Normally there is
// at most one, holding the newest points. Any other pending chunk, and one with
// a point not after every finalized chunk (it would not be the newest), is
// finalized as before, so that reads keep their order. A pending chunk without an intact point is removed, one whose
// header is not one of a chunk is left on disk and skipped. Called by openColumn only.
func (c *Column) resumePending(pending []chunkInfo) error {
	sortChunks(pending)
	sc := getScratch()
	defer putScratch(sc)
	var states []openState
	for _, p := range pending {
		st, res, err := recoverOpenChunk(c, p.path, p.minMillis, &sc.rd)
		if err != nil {
			return err
		}
		switch res {
		case recoveredEmpty:
			if err := c.pool.Remove(p.path); err != nil {
				return fmt.Errorf("tsdb: remove empty open chunk: %w", err)
			}
			continue
		case recoveredNothing:
			continue
		}
		states = append(states, st)
	}
	last := len(states) - 1
	newestBefore := minInt64
	for _, ci := range c.chunks {
		newestBefore = max(newestBefore, ci.maxMillis)
	}
	for i := range last {
		newestBefore = max(newestBefore, states[i].newest())
	}
	for i, st := range states {
		if i == last && st.minMillis > newestBefore {
			c.app.adopt(st)
			continue
		}
		other := appender{c: c}
		other.adopt(st)
		if err := other.finalizePending(); err != nil {
			return err
		}
	}
	sortChunks(c.chunks)
	return nil
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

// ensureHeadMax adds the newest time of the head to curMax, reading the head if that did not happen yet. Caller holds
// c.mu.
func (c *Column) ensureHeadMax() error {
	if c.headInMax {
		return nil
	}
	if err := c.head.ensure(); err != nil {
		return err
	}
	if hmax, ok := c.head.maxTS(); ok && (!c.haveCurMax || hmax > c.curMax) {
		c.curMax = hmax
		c.haveCurMax = true
	}
	c.headInMax = true
	return nil
}

// touch records the time of a write for ReleaseIdle.
func (c *Column) touch() {
	if c.db.opts.IdleAfter > 0 {
		c.lastWrite.Store(c.db.opts.Now().UnixNano())
	}
}

// putRouted dispatches a write to the append fast path or the head based on
// monotonicity, holding c.mu for the whole operation.
func (c *Column) putRouted(ts, val int64, str string) error {
	c.mu.Lock()
	if c.dropped {
		c.mu.Unlock()
		return errColumnGone
	}
	if err := c.ensureHeadMax(); err != nil {
		c.mu.Unlock()
		return err
	}
	c.touch()
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
//
// It does not take the lock of the column, so that it does not wait for a compaction; a dropped column has a closed
// head, which refuses it.
func (c *Column) Delete(ts int64) error {
	c.touch()
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
	// tombstone every currently-known ts in range (head, sealed and the open block) so reads are masked immediately;
	// unknown future ts in range are not affected.
	var seen []int64
	if err := c.mergedRange(min, max, func(ts, _ []int64, _ []string) bool {
		seen = append(seen, ts...)
		return true
	}); err != nil {
		return err
	}
	c.touch()
	for _, t := range seen {
		if err := c.head.tombstone(t); err != nil {
			return err
		}
	}
	c.enforceHeadBound()
	return nil
}

// DropBefore removes every chunk file whose newest point is older than ts and
// returns how many. It is the cheap way of a retention, because it only reads the
// names of the files and never opens one: a chunk is only ever removed as a
// whole, so a chunk with any point at or after ts stays completely. That includes
// the open pending chunk and its points in memory, which go only when all of them
// are older than ts.
//
// Points in the head (out-of-order writes and overwrites not yet folded into a
// chunk) stay until a later flush has moved them to a chunk. Reading and appending
// behave as before afterwards. Safe for concurrent use with writers and readers.
func (c *Column) DropBefore(ts int64) (removed int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dropped {
		return 0, errColumnGone
	}
	if pendingRemoved, err := c.app.dropIfBefore(ts); err != nil {
		return 0, err
	} else if pendingRemoved {
		removed++
	}

	kept := make([]chunkInfo, 0, len(c.chunks))
	for i, ci := range c.chunks {
		if ci.maxMillis >= ts {
			kept = append(kept, ci)
			continue
		}
		if err := c.pool.Remove(ci.path); err != nil {
			// the chunks from here on are still there, the removed ones must not be read again
			c.chunks = append(kept, c.chunks[i:]...)
			return removed, fmt.Errorf("tsdb: remove chunk: %w", err)
		}
		removed++
	}
	c.chunks = kept
	return removed, nil
}

// markDropped makes this handle fail from now on and forgets its open files, whose
// names a new column may use. Each file in the pool is evicted, because the pool
// would otherwise hand a new column the descriptor of a deleted file. The caller
// removes the files afterwards.
func (c *Column) markDropped() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dropped = true
	c.chunks = nil
	c.app = appender{c: c}
	if c.dict != nil {
		c.dict.abandon()
	}
	entries, err := listDir(c.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		c.pool.Evict(filepath.Join(c.dir, e.Name()))
	}
	return c.head.close()
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
		if !c.dropped && c.head.len() >= c.db.opts.MaxHeadPoints {
			// the point of the put is in the WAL already; a compaction, which fails, e.g. on a block of a later
			// version, removes nothing, and the next Flush or Close reports it again
			if err := c.compactLocked(); err != nil {
				slog.Error("tsdb: compaction failed", "bucket", c.bucket, "column", c.name, "err", err)
			}
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
	if c.head.ensure() != nil {
		return false
	}
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
// the sealed blocks of the appender's still-open pending chunk when they overlap the range. Its blocks are the newest
// on disk, so they are scanned after the finalized chunks; the open block (newer still) is emitted separately by the
// caller. The pointers into c.chunks let a read keep what it learns of a chunk (where its blocks begin).
func (c *Column) chunksForRead(min, max int64, sc *readScratch) []*chunkInfo {
	out := sc.chunks[:0]
	for i := range c.chunks {
		ci := &c.chunks[i]
		if ci.maxMillis < min || ci.minMillis > max {
			continue
		}
		out = append(out, ci)
	}
	if pc, ok := c.app.pendingChunk(); ok && !(pc.maxMillis < min || pc.minMillis > max) {
		sc.pc = pc
		out = append(out, &sc.pc)
	}
	sc.chunks = out
	return out
}

// scanBlocks calls fn with every intact block of ci that overlaps [qmin, qmax], in ascending order. It reads only
// these blocks and a little around them: a chunk whose blocks take more than chunkSearchFrom and which begins before
// qmin is searched for the first of them (frameReader.search), and the scan stops at the first block after qmax.
// cont is false when fn stopped it. A chunk whose header is not one of a chunk is skipped, as is a corrupt block; a
// chunk of a later version is errUnknownFormat. fn gets each block at its offset at, checked against its CRC and still
// masked (see decodeFrame).
func (c *Column) scanBlocks(ci *chunkInfo, qmin, qmax int64, sc *readScratch, fn func(at int64, frame []byte, h blockHeader) bool) (cont bool, err error) {
	if ci.bad {
		return true, nil
	}
	if ci.dataOff == 0 {
		head, err := readChunkHeader(c.pool, ci.path)
		if errors.Is(err, errCorruptChunk) {
			ci.bad = true // not a chunk: skipped as before, its blocks reported at the header
			c.reportDamaged(ci.path, 0)
			return true, nil
		}
		if err != nil {
			return false, err
		}
		ci.dataOff, ci.salt, ci.masked = head.dataOff, head.salt, head.masked
	}
	end := ci.sizeBytes
	if end <= 0 {
		if end, err = statSize(c.pool, ci.path); err != nil {
			return false, fmt.Errorf("tsdb: stat chunk: %w", err)
		}
		ci.sizeBytes = end
	}
	r := &sc.rd
	r.reset(c, ci.path, ci.dataOff, end, false)
	// only a masked chunk is searched: in an unmasked one (version 1), the bytes of a value could look like a block
	if ci.masked && end-ci.dataOff > chunkSearchFrom && ci.minMillis < qmin {
		from, err := r.search(ci.dataOff, end, qmin, ci.minMillis, ci.maxMillis)
		if err != nil {
			return false, err
		}
		r.pos, r.batch = from, readerFirstBatch
	}
	for {
		kind, at, frame, err := r.next(frameBlock)
		if err != nil {
			return false, err
		}
		if kind == frameNone {
			return true, nil
		}
		h, _ := parseBlockHeader(frame)
		if h.maxMillis < qmin {
			continue
		}
		if h.minMillis > qmax {
			return true, nil
		}
		if !fn(at, frame, h) {
			return false, nil
		}
	}
}

// decodeFrame unmasks the intact block frame of ci at offset at in place and decodes it into bd.
func (c *Column) decodeFrame(ci *chunkInfo, at int64, frame []byte, h blockHeader, bd *blockData) error {
	unmaskBlock(frame, h, ci.salt, at)
	return decodeBlockData(frame, h, c.schema.Scheme, bd)
}

// damagedPlace is the place of a damaged block: its column, the name of its chunk file and its offset in it.
type damagedPlace struct {
	bucket, column, file string
	offset               int64
}

// reportDamaged writes the place of a block, which a read skips because it fails its CRC or does not decode, to the
// log once until the DB is opened again: a later read of the same place stays silent. A chunk, which gets a new name
// when it is sealed, is a new place. The offset 0 is the header of a chunk, which is none, so that every block of it
// is skipped.
func (c *Column) reportDamaged(path string, offset int64) {
	p := damagedPlace{bucket: c.bucket, column: c.name, file: filepath.Base(path), offset: offset}
	if _, seen := c.db.damaged.LoadOrStore(p, struct{}{}); seen {
		return
	}
	slog.Warn("tsdb: damaged block skipped", "bucket", p.bucket, "column", p.column, "file", p.file, "offset", p.offset)
}

// trim returns the points of bd within [min, max] as the slices a read hands out: vals for the numeric schemes, strs
// for the string scheme.
func trim(bd *blockData, min, max int64, numeric bool) (ts, vals []int64, strs []string) {
	lo, hi := 0, len(bd.ts)
	for lo < hi && bd.ts[lo] < min {
		lo++
	}
	for hi > lo && bd.ts[hi-1] > max {
		hi--
	}
	if numeric {
		return bd.ts[lo:hi], bd.vals[lo:hi], nil
	}
	return bd.ts[lo:hi], nil, bd.strs[lo:hi]
}

// mergedRange yields every non-deleted point with min <= ts <= max in ascending
// timestamp order, merging sealed chunks and the open block with the head (head wins). It decodes
// block by block and calls fn with parallel slices; fn returns false to stop.
// For enum columns vals carries dictionary ids; string resolution is the
// caller's job. For string columns strs is populated and vals is nil.
//
// This is the single read primitive shared by the public API and DeleteRange.
//
// Allocation profile: the buffers of a read come from a pool shared by all columns and go back when it ends. The
// common read (the head does not overlap the range) streams decoded blocks straight to fn without copying.
func (c *Column) mergedRange(min, max int64, fn func(ts []int64, vals []int64, strs []string) bool) error {
	// Hold c.mu for the whole read so chunk files cannot be removed by a
	// concurrent flush/compaction while we scan them. Compaction is off the
	// write path and infrequent; correctness is prioritized here. A future
	// optimization can replace this with per-chunk refcounting for lock-free
	// reads over immutable files.
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dropped {
		return errColumnGone
	}
	if err := c.head.ensure(); err != nil {
		return err
	}
	sc := getScratch()
	defer putScratch(sc)
	scheme := c.schema.Scheme
	numeric := scheme != SchemeString
	chunks := c.chunksForRead(min, max, sc)
	ob := &sc.ob
	if lo, ok := c.app.openBlockMin(); ok && lo <= max && c.app.newest() >= min {
		if err := c.app.loadOpenBlock(ob, &sc.rd); err != nil {
			return err
		}
	}

	// Fast path: the head does not overlap the query range. This is the dominant read case. Each decoded block goes
	// directly to fn with no per-point copy.
	if !c.head.hasRange(min, max) {
		stop := false
		for _, ci := range chunks {
			var unknown error
			cont, err := c.scanBlocks(ci, min, max, sc, func(at int64, frame []byte, h blockHeader) bool {
				if err := c.decodeFrame(ci, at, frame, h, &sc.bd); err != nil {
					if errors.Is(err, errUnknownFormat) {
						unknown = err
						return false
					}
					c.reportDamaged(ci.path, at)
					return true // skip corrupt block
				}
				ts, vals, strs := trim(&sc.bd, min, max, numeric)
				if len(ts) > 0 && !fn(ts, vals, strs) {
					stop = true
					return false
				}
				return true
			})
			if err = errors.Join(err, unknown); err != nil {
				return err
			}
			if !cont || stop {
				return nil
			}
		}
		if ts, vals, strs := trim(ob, min, max, numeric); len(ts) > 0 {
			fn(ts, vals, strs)
		}
		return nil
	}

	// Slow path: the head overlaps; its entries go between the points in time order, an entry at the time of a point
	// replaces it, a tombstone removes it.
	m := merger{he: c.head.snapshot(min, max), numeric: numeric, out: &sc.out, fn: fn}
	for _, ci := range chunks {
		var unknown error
		cont, err := c.scanBlocks(ci, min, max, sc, func(at int64, frame []byte, h blockHeader) bool {
			if err := c.decodeFrame(ci, at, frame, h, &sc.bd); err != nil {
				if errors.Is(err, errUnknownFormat) {
					unknown = err
					return false
				}
				c.reportDamaged(ci.path, at)
				return true // skip corrupt block
			}
			return m.block(trim(&sc.bd, min, max, numeric))
		})
		if err = errors.Join(err, unknown); err != nil {
			return err
		}
		if !cont {
			return nil
		}
	}
	if !m.block(trim(ob, min, max, numeric)) {
		return nil
	}
	m.rest()
	return nil
}

// merger merges the entries of the head into the points of a read, which come block by block in ascending order.
type merger struct {
	he      []headEntry // ascending
	next    int
	numeric bool
	out     *blockData
	fn      func(ts []int64, vals []int64, strs []string) bool
}

func (m *merger) block(ts, vals []int64, strs []string) bool {
	o := m.out
	o.ts, o.vals, o.strs = o.ts[:0], o.vals[:0], o.strs[:0]
	for i, t := range ts {
		for m.next < len(m.he) && m.he[m.next].ts < t {
			m.emitHead()
		}
		if m.next < len(m.he) && m.he[m.next].ts == t {
			m.emitHead() // the head wins
			continue
		}
		o.ts = append(o.ts, t)
		if m.numeric {
			o.vals = append(o.vals, vals[i])
		} else {
			o.strs = append(o.strs, strs[i])
		}
	}
	return m.flush()
}

func (m *merger) emitHead() {
	e := m.he[m.next]
	m.next++
	if e.tombstone {
		return
	}
	m.out.ts = append(m.out.ts, e.ts)
	if m.numeric {
		m.out.vals = append(m.out.vals, e.val)
	} else {
		m.out.strs = append(m.out.strs, e.str)
	}
}

// rest emits the entries of the head after the last point.
func (m *merger) rest() bool {
	o := m.out
	o.ts, o.vals, o.strs = o.ts[:0], o.vals[:0], o.strs[:0]
	for m.next < len(m.he) {
		m.emitHead()
	}
	return m.flush()
}

func (m *merger) flush() bool {
	o := m.out
	if len(o.ts) == 0 {
		return true
	}
	if m.numeric {
		return m.fn(o.ts, o.vals, nil)
	}
	return m.fn(o.ts, nil, o.strs)
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

func sealedVal(bd *blockData, i int) int64 {
	return bd.vals[i]
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
