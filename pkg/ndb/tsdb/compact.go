// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/pkg/ndb"
)

// compactor runs flush/compaction of columns off the write path. Columns are
// enqueued when their head crosses a threshold; a single background goroutine
// drains the queue so writers never block on the heavy chunk rewrite. Each
// column is processed at most once per pending signal (deduplicated).
type compactor struct {
	db *DB

	mu      sync.Mutex
	pending map[*Column]struct{}
	wake    chan struct{}
	done    chan struct{}
	stopped bool
	wg      sync.WaitGroup
}

func newCompactor(db *DB) *compactor {
	c := &compactor{
		db:      db,
		pending: make(map[*Column]struct{}),
		wake:    make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
	c.wg.Add(1)
	go c.loop()
	return c
}

func (c *compactor) enqueue(col *Column) {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	c.pending[col] = struct{}{}
	c.mu.Unlock()
	select {
	case c.wake <- struct{}{}:
	default:
	}
}

func (c *compactor) loop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.done:
			return
		case <-c.wake:
			c.drain()
		}
	}
}

func (c *compactor) drain() {
	for {
		c.mu.Lock()
		var col *Column
		for k := range c.pending {
			col = k
			delete(c.pending, k)
			break
		}
		c.mu.Unlock()
		if col == nil {
			return
		}
		col.mu.Lock()
		// re-check under lock: the head may have been reset by a concurrent
		// synchronous Flush, and the column may have been dropped meanwhile.
		if !col.dropped && col.needsFlush() {
			if err := col.compactLocked(); err != nil {
				slog.Error("tsdb: compaction failed", "bucket", col.bucket, "column", col.name, "err", err)
			}
		}
		col.mu.Unlock()
	}
}

// stop drains once more and terminates the background goroutine.
func (c *compactor) stop() {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	c.stopped = true
	c.mu.Unlock()
	close(c.done)
	c.wg.Wait()
}

// ---- folding the head into the chunks ----

const (
	newSuffix      = ".new"        // a file written under another name, which a rename puts into place
	compactLogName = "compact.log" // the commit of a compaction: what it replaces
)

// compactManifest is the content of compact.log: the chunks a compaction removes, the chunks it renames into place
// and its new WAL, by their names in the column directory. The new files have names of their own, which no other
// compaction uses, so that a manifest, which comes back after a power failure, renames nothing but its own files.
type compactManifest struct {
	Remove []string    `json:"remove"`
	Rename [][2]string `json:"rename"`
	// WAL is the new WAL and WALSize the size of the old one when the new one was written: what came behind it was
	// written meanwhile and goes behind the new one.
	WAL     string `json:"wal"`
	WALSize int64  `json:"walSize"`
}

var compactSeq atomic.Uint64

// compactID names the files of one compaction.
func compactID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + strconv.FormatUint(compactSeq.Add(1), 36)
}

// newChunk is a chunk written by a compaction under a name of its own, which the commit renames to final.
type newChunk struct {
	tmp, final string
	info       chunkInfo
}

// flushLocked makes all buffered/append data durable and folds the out-of-order
// head into sealed chunks, for Column.Flush. Caller holds c.mu. Two parts:
//
//  1. The pending chunk is finalized: its open block is sealed and it is renamed to its final name.
//  2. If the head holds out-of-order corrections, the overlapping chunks are
//     rewritten to physically apply overwrites/tombstones, then the head resets.
func (c *Column) flushLocked() error { return c.foldLocked(true) }

// compactLocked is flushLocked without finalizing the pending chunk, for
// DB.Close and for the compaction of the head, so that only the split policy
// (and an explicit Column.Flush) ends a chunk, and the next open continues it.
// The head is folded as in flushLocked. If it reaches into the pending chunk,
// that chunk is rewritten with it, its open block included, and the newest of the rewritten chunks
// stays open as the new pending chunk; otherwise the pending chunk and its open block are not touched.
//
// Caller holds c.mu.
func (c *Column) compactLocked() error { return c.foldLocked(false) }

// foldLocked is flushLocked (finalize) or compactLocked (!finalize). A rewrite does not lose a point at any byte of
// a crash: see runCompactions.
func (c *Column) foldLocked(finalize bool) error {
	k, err := c.prepareFold(finalize)
	if err != nil || k == nil {
		return err
	}
	return runCompactions(c.pool, []*compaction{k})
}

// compaction is the rewrite of the chunks of one column with its head, written under names of its own and ready to
// be committed (see runCompactions).
type compaction struct {
	c              *Column
	news           []newChunk
	open           *openState
	rewritePending bool
	hmin, hmax     int64
	fold           *headFold
	m              compactManifest
	logNew         string
	committed      bool
}

// prepareFold writes what a compaction of the head of c puts in place of the chunks it touches: the new chunks, the
// new WAL and the manifest, all under names ending in .new, without syncing them. It returns nil if the head is
// empty. With finalize the pending chunk is finalized first. Caller holds c.mu until the compaction ended.
func (c *Column) prepareFold(finalize bool) (*compaction, error) {
	if c.dropped {
		return nil, errColumnGone
	}
	if err := c.head.ensure(); err != nil {
		return nil, err
	}
	if finalize {
		if err := c.app.finalizePending(); err != nil {
			return nil, err
		}
	}

	// fold the out-of-order head, if any.
	headEntries := c.head.snapshot(minInt64, maxInt64)
	if len(headEntries) == 0 {
		return nil, nil
	}

	// Determine the time span the head touches, then rewrite the union of that
	// span with any overlapping sealed chunks into fresh chunk files.
	hmin, hmax := headEntries[0].ts, headEntries[len(headEntries)-1].ts
	overlap := c.chunksOverlapping(hmin, hmax)
	// A pending chunk is only left here for compactLocked. It holds the newest
	// points, so the head reaches into it exactly when hmax is not before it.
	rewritePending := c.app.pendingPath != "" && hmax >= c.app.pendingMin

	// Expand the rewrite window to the full extent of the overlapping chunks so
	// that points inside those chunks but outside the head's span are carried
	// over rather than dropped when the old chunk files are replaced.
	for _, ci := range overlap {
		hmin, hmax = min(hmin, ci.minMillis), max(hmax, ci.maxMillis)
	}
	if rewritePending {
		hmin, hmax = min(hmin, c.app.pendingMin), max(hmax, c.app.newest())
	}

	// Build the merged, ordered, de-duplicated point set across overlapping
	// sealed chunks, the open block + head (head wins, tombstones drop).
	sc := getScratch()
	defer putScratch(sc)
	merged := map[int64]mergedPoint{}
	take := func(bd *blockData) {
		for i, t := range bd.ts {
			merged[t] = sealedPoint(bd, c.schema.Scheme, i)
		}
	}
	scan := func(ci *chunkInfo) error {
		_, err := c.scanBlocks(ci, minInt64, maxInt64, sc, func(at int64, frame []byte, h blockHeader) bool {
			if c.decodeFrame(ci, at, frame, h, &sc.bd) == nil {
				take(&sc.bd)
			}
			return true
		})
		return err
	}
	var replaced []string
	for i := range overlap {
		if err := scan(&overlap[i]); err != nil {
			return nil, err
		}
		replaced = append(replaced, overlap[i].path)
	}
	if rewritePending {
		if pc, ok := c.app.pendingChunk(); ok {
			if err := scan(&pc); err != nil {
				return nil, err
			}
		}
		if err := c.app.loadOpenBlock(&sc.ob, &sc.rd); err != nil {
			return nil, err
		}
		take(&sc.ob)
		replaced = append(replaced, c.app.pendingPath)
	}
	for _, e := range headEntries {
		if e.tombstone {
			delete(merged, e.ts)
			continue
		}
		merged[e.ts] = mergedPoint{val: e.val, str: e.str}
	}
	order := make([]int64, 0, len(merged))
	for t := range merged {
		order = append(order, t)
	}
	slices.Sort(order)

	// write new chunk files for the merged range, respecting the split policy;
	// the newest of them stays open in place of a rewritten pending chunk
	k := &compaction{c: c, rewritePending: rewritePending, hmin: hmin, hmax: hmax}
	id := compactID()
	var err error
	k.news, k.open, err = c.writeChunks(id, order, merged, rewritePending)
	if err != nil {
		k.abort()
		return nil, err
	}
	if k.fold, err = c.head.foldBegin(foldMark(headEntries), filepath.Join(c.dir, "head-"+id+".wal"+newSuffix)); err != nil {
		k.abort()
		return nil, err
	}
	final := map[string]bool{}
	k.m = compactManifest{WAL: filepath.Base(k.fold.newPath), WALSize: k.fold.oldSize}
	for _, n := range k.news {
		final[filepath.Base(n.final)] = true
		k.m.Rename = append(k.m.Rename, [2]string{filepath.Base(n.tmp), filepath.Base(n.final)})
	}
	for _, p := range replaced {
		if name := filepath.Base(p); !final[name] {
			k.m.Remove = append(k.m.Remove, name)
		}
	}
	data, err := json.Marshal(k.m)
	if err == nil {
		k.logNew = filepath.Join(c.dir, compactLogName+newSuffix)
		err = c.pool.WriteFile(k.logNew, data)
	}
	if err != nil {
		k.abort()
		return nil, fmt.Errorf("tsdb: compaction manifest: %w", err)
	}
	return k, nil
}

// toSync are the files the compaction wrote and the directory of the column.
func (k *compaction) toSync() []string {
	var paths []string
	for _, n := range k.news {
		paths = append(paths, n.tmp)
	}
	if k.fold != nil && k.fold.newSize > 0 {
		paths = append(paths, k.fold.newPath)
	}
	return append(paths, k.logNew, k.c.dir)
}

// abort removes the files of a compaction, which did not commit.
func (k *compaction) abort() {
	for _, n := range k.news {
		_ = k.c.pool.Remove(n.tmp)
	}
	if k.fold != nil {
		k.fold.abort()
	}
	if k.logNew != "" {
		_ = k.c.pool.Remove(k.logNew)
	}
}

// commit renames compact.log.new to compact.log: from now on the next open finishes the compaction.
func (k *compaction) commit() error {
	if err := k.c.pool.Rename(k.logNew, filepath.Join(k.c.dir, compactLogName)); err != nil {
		k.abort()
		return fmt.Errorf("tsdb: commit compaction: %w", err)
	}
	k.committed = true
	return nil
}

// apply removes the replaced chunks, renames the new files into place and takes the new state into the column.
func (k *compaction) apply() error {
	c := k.c
	if err := c.applyChunks(k.m); err != nil {
		return fmt.Errorf("tsdb: apply compaction: %w", err)
	}
	if err := k.fold.finish(); err != nil {
		return fmt.Errorf("tsdb: apply compaction: %w", err)
	}
	if k.rewritePending {
		c.app.forgetPending()
		if k.open != nil {
			c.app.adopt(*k.open)
		}
	}
	kept := make([]chunkInfo, 0, len(c.chunks))
	for _, ci := range c.chunks {
		if ci.maxMillis < k.hmin || ci.minMillis > k.hmax {
			kept = append(kept, ci)
		}
	}
	for _, n := range k.news {
		if !n.info.pending {
			kept = append(kept, n.info)
		}
	}
	sortChunks(kept)
	c.chunks = kept
	return nil
}

// end removes compact.log.
func (k *compaction) end() error {
	if err := k.c.pool.Remove(filepath.Join(k.c.dir, compactLogName)); err != nil {
		return fmt.Errorf("tsdb: end compaction: %w", err)
	}
	return nil
}

// runCompactions puts prepared compactions in place, of one column or, for DB.Close, of many, whose syncs it bundles:
//
//  1. The new chunks, the new WALs and the manifests compact.log.new of all of them and their directories are synced
//     (SyncAll: one flush of the drive for all).
//  2. Each manifest is renamed to compact.log, which commits its compaction; the directories are synced.
//  3. Each compaction removes the chunks it replaces and renames its new chunks into place; what writers put into its
//     old WAL meanwhile goes behind the new one, which is renamed over it (compaction.apply). The directories are
//     synced.
//  4. Each compact.log is removed.
//
// So no old chunk and no old WAL goes before the new files are on the disk, and no compact.log before what it names
// is done. Before 2 a crash leaves the old state and files ending in .new, which the next open removes; after it the
// next open finishes 3 (redoCompaction). Writers of the head are held up only by 3. The callers hold the locks of
// the columns.
func runCompactions(pool *ndb.FilePool, ks []*compaction) error {
	var paths, dirs []string
	for _, k := range ks {
		paths = append(paths, k.toSync()...)
		dirs = append(dirs, k.c.dir)
	}
	if err := pool.SyncAll(paths...); err != nil {
		parallel(len(ks), func(i int) { ks[i].abort() })
		return fmt.Errorf("tsdb: sync compaction: %w", err)
	}
	errs := make([]error, len(ks))
	parallel(len(ks), func(i int) { errs[i] = ks[i].commit() })
	if err := pool.SyncAll(dirs...); err != nil {
		return fmt.Errorf("tsdb: sync compaction: %w", errors.Join(append(errs, err)...))
	}
	applied := make([]bool, len(ks))
	parallel(len(ks), func(i int) {
		if ks[i].committed {
			if errs[i] = ks[i].apply(); errs[i] == nil { // else the next open finishes it
				applied[i] = true
			}
		}
	})
	if err := pool.SyncAll(dirs...); err != nil {
		return fmt.Errorf("tsdb: sync compaction: %w", errors.Join(append(errs, err)...))
	}
	parallel(len(ks), func(i int) {
		if applied[i] {
			errs[i] = ks[i].end()
		}
	})
	return errors.Join(errs...)
}

// parallel runs fn for 0..n-1 on up to 16 goroutines, so that the file operations of many columns, which wait for the
// file system, overlap.
func parallel(n int, fn func(i int)) {
	if n == 1 {
		fn(0)
		return
	}
	var next atomic.Int64
	var wg sync.WaitGroup
	for range min(n, 16) {
		wg.Go(func() {
			for i := int(next.Add(1) - 1); i < n; i = int(next.Add(1) - 1) {
				fn(i)
			}
		})
	}
	wg.Wait()
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

// writeChunks writes the ordered points into new chunk files under names of their own, sealing per the
// split policy. With keepOpen the last chunk is not finalized: its final name is the one of a pending chunk, and open
// is its state, for the appender to continue.
func (c *Column) writeChunks(id string, order []int64, merged map[int64]mergedPoint, keepOpen bool) (news []newChunk, open *openState, err error) {
	i := 0
	var bd blockData
	var frameBuf []byte
	for i < len(order) {
		startTs := order[i]
		tmp := filepath.Join(c.dir, fmt.Sprintf("compact-%s-%d%s", id, len(news), newSuffix))
		salt := newSalt()
		if _, err := c.pool.WriteAt(tmp, chunkHeader(chunkSlot{seq: 1, end: chunkDataOff}, salt), 0); err != nil {
			return news, nil, fmt.Errorf("tsdb: write chunk header: %w", err)
		}
		news = append(news, newChunk{tmp: tmp})
		fileOff := chunkDataOff
		var points int64
		lastTs := startTs

		flushBlock := func() error {
			if bd.count() == 0 {
				return nil
			}
			f, err := encodeBlock(frameBuf[:0], &bd, c.schema.Scheme, c.db.opts.Compress)
			if err != nil {
				return err
			}
			frameBuf = f
			maskBlock(f, salt, fileOff)
			if _, err := c.pool.WriteAt(tmp, f, fileOff); err != nil {
				return err
			}
			fileOff += int64(len(f))
			bd.ts, bd.vals, bd.strs = bd.ts[:0], bd.vals[:0], bd.strs[:0]
			return nil
		}

		for i < len(order) {
			ts := order[i]
			stats := ChunkStats{
				SizeBytes: fileOff, Points: points,
				MinMillis: startTs, MaxMillis: lastTs, NextMillis: ts,
			}
			if points > 0 && c.db.opts.Split(stats) {
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
					return news, nil, err
				}
			}
		}
		if err := flushBlock(); err != nil {
			return news, nil, err
		}
		if err := writeSlot(c.pool, tmp, chunkSlot{seq: 2, end: fileOff, points: points, max: lastTs}); err != nil {
			return news, nil, err
		}

		n := &news[len(news)-1]
		if keepOpen && i == len(order) {
			// the newest chunk stays pending under its <min>_.tsb name
			n.final = filepath.Join(c.dir, pendingChunkName(startTs))
			n.info = chunkInfo{path: n.final, minMillis: startTs, maxMillis: lastTs, pending: true, sizeBytes: fileOff,
				dataOff: chunkDataOff, salt: salt, masked: true}
			return news, &openState{path: n.final, minMillis: startTs, dataOff: chunkDataOff, salt: salt, masked: true, slotSeq: 2,
				blocksEnd: fileOff, blocksPts: points, blocksMax: lastTs, fileEnd: fileOff}, nil
		}
		n.final = filepath.Join(c.dir, finalChunkName(startTs, lastTs))
		n.info = chunkInfo{path: n.final, minMillis: startTs, maxMillis: lastTs, sizeBytes: fileOff, dataOff: chunkDataOff,
			salt: salt, masked: true}
	}
	return news, nil, nil
}

// applyChunks removes the chunks of m and renames its new chunks into place. It can run again after a crash in its
// middle: a removed chunk or a renamed file is no more there.
func (c *Column) applyChunks(m compactManifest) error {
	for _, name := range m.Remove {
		if err := c.pool.Remove(filepath.Join(c.dir, name)); err != nil {
			return err
		}
	}
	for _, r := range m.Rename {
		if err := c.renameIfThere(r[0], r[1]); err != nil {
			return err
		}
	}
	return nil
}

func (c *Column) renameIfThere(from, to string) error {
	src := filepath.Join(c.dir, from)
	if _, err := c.pool.Stat(src); err != nil {
		if isNotExist(err) {
			return nil // renamed before
		}
		return err
	}
	return c.pool.Rename(src, filepath.Join(c.dir, to))
}

// redoCompaction finishes a compaction, which committed by compact.log before a crash.
func (c *Column) redoCompaction() error {
	logPath := filepath.Join(c.dir, compactLogName)
	data, err := c.pool.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("tsdb: read compaction manifest: %w", err)
	}
	var m compactManifest
	if err := json.Unmarshal(data, &m); err != nil {
		// it was renamed into place only after it was synced, so this is not one of ours
		slog.Error("tsdb: unreadable compaction manifest, ignored", "dir", c.dir, "err", err)
	} else if err := c.redoManifest(m); err != nil {
		return fmt.Errorf("tsdb: finish compaction: %w", err)
	}
	if err := errors.Join(c.pool.SyncAll(c.dir), c.pool.Remove(logPath)); err != nil {
		return fmt.Errorf("tsdb: end compaction: %w", err)
	}
	return nil
}

// redoManifest is the rest of a committed compaction after a crash: its chunks, then its WAL, behind which goes what
// the old one holds beyond WALSize.
func (c *Column) redoManifest(m compactManifest) error {
	if err := c.applyChunks(m); err != nil {
		return err
	}
	if m.WAL == "" {
		return nil
	}
	walNew := filepath.Join(c.dir, m.WAL)
	newSize, err := statSize(c.pool, walNew)
	if err != nil {
		if isNotExist(err) {
			return nil // renamed before
		}
		return err
	}
	walPath := filepath.Join(c.dir, headWALName)
	oldSize, err := statSize(c.pool, walPath)
	if err != nil && !isNotExist(err) {
		return err
	}
	if err := appendWALTail(c.pool, walPath, m.WALSize, oldSize, walNew, newSize); err != nil {
		return err
	}
	return c.pool.Rename(walNew, walPath)
}
