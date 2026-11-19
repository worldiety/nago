// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sync/atomic"
	"time"

	"github.com/google/btree"
)

// lockedTail is the part of the log, which a compaction copies while it holds the write lock. It copies the larger
// part of the log, which arrives meanwhile, without the lock before.
const lockedTail = 256 * 1024

// maybeCompact starts a compaction in the background, if the garbage has reached the limit of the options, see
// [state.plan]. The caller holds the write lock.
func (db *DB) maybeCompact() {
	if db.closed || db.broken != nil || db.compacting.Load() {
		return
	}

	due, logOnly := db.st.plan(db.opts)
	if !due {
		return
	}

	if !db.retryAt.IsZero() && time.Now().Before(db.retryAt) {
		return
	}

	db.compacting.Store(true)
	db.bg.Add(1)
	go func() {
		defer db.bg.Done()

		err := db.compact(logOnly)
		db.mu.Lock()
		defer db.mu.Unlock()

		db.compacting.Store(false)
		if err != nil {
			if !errors.Is(err, ErrClosed) {
				slog.Error("tdb compaction failed", "dir", db.dir, "err", err)
				db.retryAt = time.Now().Add(time.Minute)
			}

			return
		}

		// the writes while compacting may have left enough garbage for the next compaction
		db.maybeCompact()
	}()
}

// Compact writes the entries into a new compacted file and the changes, which arrive meanwhile, into a new log, which
// then replace the former files. Reads and writes go on while it runs, except for a short moment at the end. Readers
// of values and iterators, which have started before, keep reading the former files. A DB compacts by itself, when
// the garbage of its files grows too large, see [Options]. Then, it may compact the log alone, which keeps the
// compacted file.
func (db *DB) Compact() error {
	return db.compact(false)
}

func (db *DB) compact(logOnly bool) error {
	db.compactMu.Lock()
	defer db.compactMu.Unlock()

	if logOnly {
		return db.compactLog()
	}

	return db.compactAll()
}

// compactAll compacts both files, the caller holds compactMu.
func (db *DB) compactAll() (err error) {
	start := time.Now()

	// take a snapshot of the trees, every later change is in the log behind cut
	db.mu.Lock()
	if err := db.writable(); err != nil {
		db.mu.Unlock()
		return err
	}

	old := db.st
	snapTx := db.tx
	cut := old.log.size.Load()
	names := make([]string, 0, len(old.buckets))
	for name, t := range old.buckets {
		if t.Len() > 0 {
			names = append(names, name)
		}
	}

	slices.Sort(names)
	trees := make([]*btree.BTreeG[item], len(names))
	for i, name := range names {
		trees[i] = old.buckets[name].Clone()
	}

	garbageBefore := old.garbage()
	db.mu.Unlock()

	next := &state{buckets: make(map[string]*btree.BTreeG[item], len(names)), snapTx: snapTx}

	// write the new compacted file and replace the former one, which is consistent with the former log, because a
	// load skips its records up to snapTx
	snapNext := db.snapPath + nextSuffix
	snapF, err := openNext(snapNext)
	if err != nil {
		return err
	}

	snapReplaced := false
	installed := false
	defer func() {
		if !installed {
			_ = snapF.Close()
			if !snapReplaced {
				_ = os.Remove(snapNext)
			}
		}
	}()

	snapSize, err := db.writeSnapshot(snapF, old, names, trees, next)
	if err != nil {
		return err
	}

	trees = nil
	if err := snapF.Sync(); err != nil {
		return err
	}

	next.snap = &dataFile{f: snapF}
	next.snap.size.Store(snapSize)
	if err := replaceFile(db.snapPath, snapNext); err != nil {
		if !exists(db.snapPath) {
			db.markBroken(fmt.Errorf("tdb compacted file is missing after a failed replacement: %w", err))
		}

		return fmt.Errorf("cannot replace compacted file: %w", err)
	}

	snapReplaced = true
	if err := syncDir(db.dir); err != nil {
		return err
	}

	// copy the log behind cut, without the lock as long as there is a lot of it
	logNext := db.logPath + nextSuffix
	logF, err := openNext(logNext)
	if err != nil {
		return err
	}

	defer func() {
		if !installed {
			_ = logF.Close()
			_ = os.Remove(logNext)
		}
	}()

	// The bulk of the new log is synced, otherwise a machine crash could leave it empty after the rename, while the
	// former log has long been on disk. The writes during the sync are copied afterwards, still without the lock.
	tail := &tailCopy{src: old.log, dst: bufio.NewWriterSize(logF, 1024*1024), next: next}
	copied, err := db.catchUp(tail, cut)
	if err != nil {
		return err
	}

	if err := logF.Sync(); err != nil {
		return err
	}

	durable := copied
	copied, err = db.catchUp(tail, copied)
	if err != nil {
		return err
	}

	// copy the rest and replace the log and the state, either log is consistent with the new compacted file
	lockStart := time.Now()
	if err := db.install(old, next, logF, tail, copied, durable); err != nil {
		return err
	}

	installed = true
	lockDuration := time.Since(lockStart)
	db.syncDirAfterInstall()
	db.compactions.Add(1)
	slog.Info("tdb compaction complete", "dir", db.dir, "file-size", snapSize+tail.size, "garbage-before", garbageBefore,
		"duration", time.Since(start), "locked", lockDuration)

	return nil
}

// compactLog writes the live records and the removals of the log into a new log, which replaces the former one, while
// the compacted file stays as it is. Like compactAll, it holds the write lock only for a short moment at the end. The
// caller holds compactMu.
func (db *DB) compactLog() (err error) {
	start := time.Now()

	// take a snapshot of the trees, every later change is in the log behind cut
	db.mu.Lock()
	if err := db.writable(); err != nil {
		db.mu.Unlock()
		return err
	}

	old := db.st
	cut := old.log.size.Load()
	next := &state{
		snap:        old.snap,
		snapTx:      old.snapTx,
		buckets:     make(map[string]*btree.BTreeG[item], len(old.buckets)),
		snapGarbage: old.snapGarbage,
	}

	for name, t := range old.buckets {
		next.buckets[name] = t.Clone()
	}

	garbageBefore := old.garbage()
	db.mu.Unlock()

	logNext := db.logPath + nextSuffix
	logF, err := openNext(logNext)
	if err != nil {
		return err
	}

	installed := false
	defer func() {
		if !installed {
			_ = logF.Close()
			_ = os.Remove(logNext)
		}
	}()

	// the same steps as compactAll, except that the log up to cut is filtered instead of copied
	tail := &tailCopy{src: old.log, dst: bufio.NewWriterSize(logF, 1024*1024), next: next}
	if err := tail.filter(cut, &db.closing); err != nil {
		return err
	}

	copied, err := db.catchUp(tail, cut)
	if err != nil {
		return err
	}

	if err := logF.Sync(); err != nil {
		return err
	}

	durable := copied
	copied, err = db.catchUp(tail, copied)
	if err != nil {
		return err
	}

	lockStart := time.Now()
	if err := db.install(old, next, logF, tail, copied, durable); err != nil {
		return err
	}

	installed = true
	lockDuration := time.Since(lockStart)
	db.syncDirAfterInstall()
	db.logCompactions.Add(1)
	slog.Debug("tdb log compaction complete", "dir", db.dir, "log-size", tail.size, "garbage-before", garbageBefore,
		"duration", time.Since(start), "locked", lockDuration)

	return nil
}

// catchUp copies the log behind copied without the lock, until little enough is left for the lock, and returns the
// new end of the copy.
func (db *DB) catchUp(tail *tailCopy, copied int64) (int64, error) {
	for {
		if db.closing.Load() {
			return 0, ErrClosed
		}

		end := tail.src.size.Load()
		if end-copied <= lockedTail {
			return copied, tail.dst.Flush()
		}

		if err := tail.copy(copied, end); err != nil {
			return 0, err
		}

		copied = end
	}
}

// openNext creates the file, which a compaction writes, and locks it, see lockFile.
func openNext(path string) (*os.File, error) {
	f, err := openDataFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return nil, err
	}

	if err := lockFile(f); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return nil, err
	}

	return f, nil
}

// install copies the rest of the log behind copied into the new log and replaces the state, while it holds the write
// lock. The new log holds the former one up to durable on disk. If Sync has synced more of the former log or is
// syncing it, the new log is synced as well, otherwise a machine crash may lose those records after the rename.
func (db *DB) install(old, next *state, logF *os.File, tail *tailCopy, copied, durable int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.writable(); err != nil {
		return err
	}

	if err := tail.copy(copied, old.log.size.Load()); err != nil {
		return err
	}

	if err := tail.dst.Flush(); err != nil {
		return err
	}

	if db.syncing.Load() > 0 || old.log.synced.Load() > durable {
		if err := logF.Sync(); err != nil {
			return err
		}
	}

	next.log = &dataFile{f: logF}
	next.log.size.Store(tail.size)
	if err := replaceFile(db.logPath, logF.Name()); err != nil {
		if !exists(db.logPath) {
			// the old state would append to a file, which the next Open does not read
			db.broken = fmt.Errorf("tdb log is missing after a failed replacement: %w", err)
		}

		return fmt.Errorf("cannot replace log: %w", err)
	}

	db.st = next
	db.retryAt = time.Time{}
	db.renames.Add(1)
	return nil
}

// syncDirAfterInstall syncs the directory, after install has replaced the log. Otherwise, a machine crash may bring
// back the former log, which lacks the writes since. Sync does the same, if it comes first.
func (db *DB) syncDirAfterInstall() {
	if err := db.syncRenames(); err != nil {
		// the new log is in use, thus the compaction has succeeded, and Sync tries again
		slog.Error("tdb cannot sync the directory after replacing the log", "dir", db.dir, "err", err)
	}
}

func (db *DB) markBroken(err error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.broken = err
}

// writeSnapshot writes the entries of the trees with the transaction of the next state into f, builds the trees of
// the next state and returns the size of the file.
func (db *DB) writeSnapshot(f *os.File, old *state, names []string, trees []*btree.BTreeG[item], next *state) (int64, error) {
	w := bufio.NewWriterSize(f, 1024*1024)
	var off int64
	var rec, val []byte
	var items []item
	for i, name := range names {
		items = items[:0]
		var err error
		n := 0
		trees[i].Ascend(func(it item) bool {
			if n++; n%1024 == 0 && db.closing.Load() {
				err = ErrClosed
				return false
			}

			val = slices.Grow(val[:0], int(it.len))[:it.len]
			v := old.value(it)
			if _, err = v.f.ReadAt(val, v.offset); err != nil {
				return false
			}

			var valOff int
			rec, valOff = appendRecord(rec[:0], setKeyValue, next.snapTx, name, it.key, val)
			if _, err = w.Write(rec); err != nil {
				return false
			}

			items = append(items, item{key: next.keys.add(it.key), off: uint64(off) + uint64(valOff), len: it.len})
			off += int64(len(rec))
			return true
		})

		if err != nil {
			return 0, err
		}

		next.buckets[name] = buildTree(items)
	}

	return off, w.Flush()
}

// tailCopy copies records of the former log into the new one and applies them to the next state.
type tailCopy struct {
	src  *dataFile
	dst  *bufio.Writer
	next *state
	size int64 // bytes written to the new log
}

// filter copies the records of the log up to the offset to, which the next state still refers to, and the removals of
// the entries, which are still removed, because the compacted file may hold such an entry. It drops the rest: the
// records of overwritten and removed entries, the removals followed by a new entry and the records, which the
// compacted file already contains. The next state starts with clones of the trees of the former state at offset to.
func (c *tailCopy) filter(to int64, closing *atomic.Bool) error {
	n := 0
	end, err := scanRecords(c.src.f, 0, to, func(rec record, off int64, raw []byte) error {
		if n++; n%1024 == 0 && closing.Load() {
			return ErrClosed
		}

		if rec.tx <= c.next.snapTx {
			return nil
		}

		t := c.next.buckets[unsafeString(rec.bucket)]
		var it item
		found := false
		if t != nil {
			it, found = t.Get(item{key: unsafeString(rec.key)})
		}

		if rec.kind == removeKeyValue {
			if found {
				return nil
			}

			c.next.removal(int64(len(raw)))
		} else {
			if !found || it.off != inLog|uint64(rec.valOff) {
				return nil
			}

			// the key moves into the arena of the next state, thus the blocks of the former arena can be freed
			it.key = c.next.keys.add(it.key)
			it.off = inLog | uint64(c.size+rec.valOff-off)
			t.ReplaceOrInsert(it)
		}

		if _, err := c.dst.Write(raw); err != nil {
			return err
		}

		c.size += int64(len(raw))
		return nil
	})

	if err == nil && end != to {
		err = fmt.Errorf("log ends at %d instead of %d", end, to)
	}

	return err
}

func (c *tailCopy) copy(from, to int64) error {
	if from == to {
		return nil
	}

	end, err := scanRecords(c.src.f, from, to, func(rec record, off int64, raw []byte) error {
		if _, err := c.dst.Write(raw); err != nil {
			return err
		}

		c.next.apply(rec, int64(len(raw)), rec.valOff-off+c.size)
		c.size += int64(len(raw))
		return nil
	})

	if err == nil && end != to {
		err = fmt.Errorf("log tail ends at %d instead of %d", end, to)
	}

	return err
}
