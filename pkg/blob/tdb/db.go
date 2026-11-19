// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/btree"
	"github.com/rogpeppe/go-internal/lockedfile"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xmaps"
)

var globalDirLocks = xmaps.NewConcurrentMap[string, *sync.Mutex]()

// ErrClosed is returned by writes and compactions of a closed DB.
var ErrClosed = errors.New("tdb is closed")

const (
	snapName   = "data.tdb" // the compacted file
	logName    = "tdb.wal"  // the log of the changes since the compaction
	lockName   = "tdb.lock" // held by the process, which has opened the directory
	nextSuffix = ".next"    // a file, which a compaction writes
	oldSuffix  = ".old"     // a file, which a compaction moves away on Windows
)

const (
	// DefaultGCPercent is the default of [Options.GCPercent].
	DefaultGCPercent = 100
	// DefaultMinGarbage is the default of [Options.MinGarbage].
	DefaultMinGarbage = 4 * 1024 * 1024
)

// maxCompareSize is the largest value, which Set compares with the stored one to skip writing an unchanged value.
const maxCompareSize = 128 * 1024

// Options configure the automatic compaction.
type Options struct {
	// GCPercent works like GOGC: a compaction starts, when the garbage of the files, which are the records of
	// overwritten and removed entries and the removals themselves, reaches this percentage of the live data. Thus, the
	// default of 100 lets the files grow to about twice the live data. Zero means the default, a negative value
	// disables the automatic compaction.
	GCPercent int

	// MinGarbage is the garbage, below which no automatic compaction starts, so that a small database is not compacted
	// all the time. Zero means [DefaultMinGarbage].
	MinGarbage int64
}

// Stats describe the files of a DB.
type Stats struct {
	// FileSize is the size of the compacted file and the log.
	FileSize int64
	// Garbage estimates the bytes of the files, which a compaction removes.
	Garbage int64
	// Compactions counts the compactions of both files since Open.
	Compactions uint64
	// LogCompactions counts the compactions since Open, which only rewrote the log.
	LogCompactions uint64
	// Compacting tells, whether an automatic compaction is running.
	Compacting bool
}

type IndexEntry struct {
	key string
	val Value
}

// dataFile is the compacted file or the log of a state.
type dataFile struct {
	f      *os.File
	size   atomic.Int64 // bytes of the complete records, the log grows under the write lock of the DB
	synced atomic.Int64 // bytes of the log, which Sync has synced
}

// state is a generation of the files together with the index, which refers to them. A compaction builds a new state
// and replaces the current one, while snapshots of the iterators and readers of values keep using the former one.
// Thus, the files of a state are never closed before the DB, but left to the garbage collector.
type state struct {
	snap    *dataFile
	log     *dataFile
	snapTx  uint64 // the transaction of all records of the compacted file
	buckets map[string]*btree.BTreeG[item]
	keys    keyArena

	// The garbage of the files, which is an estimate for the log, see recordSize. A compaction of the log frees its
	// garbage except for the removals.
	snapGarbage int64 // records of the compacted file, which the log has replaced or removed
	logGarbage  int64 // records of the log, which later ones have replaced or removed, and the removals
	logRemovals int64 // the removals of the log
}

func (s *state) garbage() int64 {
	return s.snapGarbage + s.logGarbage
}

// drop accounts the record of the item as garbage, which an overwrite or a removal with the transaction tx has left
// behind.
func (s *state) drop(bucket string, it item, tx uint64) {
	size := s.recordSize(bucket, it, tx)
	if it.off&inLog != 0 {
		s.logGarbage += size
	} else {
		s.snapGarbage += size
	}
}

// removal accounts a removal record of the log as garbage.
func (s *state) removal(size int64) {
	s.logGarbage += size
	s.logRemovals += size
}

// plan tells whether the garbage requires a compaction, see [Options], and whether to compact the log alone.
//
// The log holds the entries, which have changed since the last compaction of both files. Thus, entries which change
// often stay in the log, and compacting it alone does not rewrite the entries, which rarely change. It writes the live
// records and the removals of the log and frees the rest of its garbage. It is chosen, if it frees at least as much
// garbage per written byte as a compaction of both files, which writes the live data and frees all garbage, and if it
// writes at most half of the live data. Otherwise, as with evenly spread changes, the log would hold most of the live
// data and its removals for good, while the compacted file loads faster and needs less memory, see buildTree.
func (s *state) plan(opts Options) (due, logOnly bool) {
	garbage := s.garbage()
	live := s.snap.size.Load() + s.log.size.Load() - garbage
	if opts.GCPercent < 0 || garbage < opts.MinGarbage || garbage*100 < live*int64(opts.GCPercent) {
		return false, false
	}

	logFree := s.logGarbage - s.logRemovals
	logWrite := s.log.size.Load() - logFree
	return true, logFree > 0 && 2*logWrite <= live && float64(logFree)*float64(live) >= float64(garbage)*float64(logWrite)
}

func (s *state) value(it item) Value {
	f := s.snap.f
	if it.off&inLog != 0 {
		f = s.log.f
	}

	return Value{f: f, offset: int64(it.off &^ inLog), len: it.len}
}

// recordSize returns the size of the record of the item. The transaction of a log record is not kept, thus tx gives the
// length to assume.
func (s *state) recordSize(bucket string, it item, tx uint64) int64 {
	if it.off&inLog == 0 {
		tx = s.snapTx
	}

	return recordSize(tx, len(bucket), len(it.key), it.len)
}

// apply applies a record of the log of the state, whose value is located at valOff, and accounts the garbage.
func (s *state) apply(rec record, recLen int64, valOff int64) {
	bucket := unsafeString(rec.bucket)
	key := unsafeString(rec.key)
	t, ok := s.buckets[bucket]
	if rec.kind == removeKeyValue {
		s.removal(recLen)
		if ok {
			if old, removed := t.Delete(item{key: key}); removed {
				s.drop(bucket, old, rec.tx)
			}
		}

		return
	}

	if !ok {
		t = newTree()
		s.buckets[strings.Clone(bucket)] = t
	}

	if old, found := t.Get(item{key: key}); found {
		s.drop(bucket, old, rec.tx)
		key = old.key
	} else {
		key = s.keys.add(key)
	}

	t.ReplaceOrInsert(item{key: key, off: inLog | uint64(valOff), len: rec.valLen})
}

// DB keeps an index of all entries in memory and their values in two files: the compacted file, which holds the
// entries of the last compaction, and the log, which appends every change since. A compaction runs in the background,
// when the garbage of the files grows too large, see [Options].
type DB struct {
	mu      sync.RWMutex
	st      *state // guarded by mu, as well as the garbage counters of st
	tx      uint64 // guarded by mu
	closed  bool   // guarded by mu
	broken  error  // guarded by mu, set when the files on disk do not match st anymore
	retryAt time.Time
	wbuf    []byte // encodes the records of Set and Delete, guarded by mu
	cbuf    []byte // compares values in Set, guarded by mu

	dir      string
	snapPath string
	logPath  string
	opts     Options
	dirLock  *lockedfile.File
	procLock *sync.Mutex

	// Sync of the files and the directory, which install may replace meanwhile
	syncing     atomic.Int32 // the calls of Sync, which are syncing a log
	renames     atomic.Int64 // the logs, which install has moved into place
	syncedDir   atomic.Int64 // the renames, which a sync of the directory has made durable
	syncedDirMu sync.Mutex

	compactMu      sync.Mutex // one compaction at a time
	compacting     atomic.Bool
	compactions    atomic.Uint64
	logCompactions atomic.Uint64
	closing        atomic.Bool
	bg             sync.WaitGroup
}

// Open opens the database in the directory with the default [Options].
func Open(dir string) (*DB, error) {
	return OpenWithOptions(dir, Options{})
}

// OpenWithOptions opens the database in the directory, which is created if missing. Only a single DB may open a
// directory at a time, also across processes.
func OpenWithOptions(dir string, opts Options) (*DB, error) {
	if opts.GCPercent == 0 {
		opts.GCPercent = DefaultGCPercent
	}

	if opts.MinGarbage == 0 {
		opts.MinGarbage = DefaultMinGarbage
	}

	if err := os.MkdirAll(dir, 0777); err != nil && !os.IsExist(err) {
		return nil, err
	}

	mutex, _ := globalDirLocks.LoadOrStore(dir, &sync.Mutex{})
	if ok := mutex.TryLock(); !ok {
		return nil, fmt.Errorf("tdb directory is already opened by another instance: %s", dir)
	}

	// the data files are no lockedfiles, because a compaction replaces them, while readers may still use them
	dirLock, err := lockedfile.OpenFile(filepath.Join(dir, lockName), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		mutex.Unlock()
		return nil, err
	}

	db := &DB{
		dir:      dir,
		snapPath: filepath.Join(dir, snapName),
		logPath:  filepath.Join(dir, logName),
		opts:     opts,
		dirLock:  dirLock,
		procLock: mutex,
	}

	rewrite, err := db.load()
	if err != nil {
		// lockedfile panics from a finalizer when an open file becomes unreachable
		_ = dirLock.Close()
		mutex.Unlock()
		return nil, err
	}

	// a compaction reads the log from its start and would stop at the skipped records, thus they are removed right away
	if rewrite {
		if err := db.Compact(); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("cannot rewrite the log of %s: %w", dir, err)
		}
	}

	db.mu.Lock()
	db.maybeCompact()
	db.mu.Unlock()

	return db, nil
}

// load recovers an interrupted compaction, reads the compacted file and replays the log. On error, no file is left
// open. It tells whether the log must be rewritten, because it has skipped bad records, see scanFile.
func (db *DB) load() (rewrite bool, err error) {
	if err := recoverFiles(db.dir); err != nil {
		return false, err
	}

	start := time.Now()
	st := &state{buckets: map[string]*btree.BTreeG[item]{}}
	snapF, err := openLocked(db.snapPath)
	if err != nil {
		return false, err
	}

	logF, err := openLocked(db.logPath)
	if err != nil {
		_ = snapF.Close()
		return false, err
	}

	defer func() {
		if err != nil {
			_ = logF.Close()
			_ = snapF.Close()
		}
	}()

	// only now, a former version cannot be writing them anymore
	if err := removeTmpFiles(db.dir); err != nil {
		return false, err
	}

	info, err := snapF.Stat()
	if err != nil {
		return false, err
	}

	snapSize := info.Size()

	// The records of a bucket follow each other in ascending order, thus a bucket is built at once, which fills the
	// nodes of its tree, see buildTree.
	var bucket string
	var items []item
	haveBucket := false
	flush := func() {
		if !haveBucket {
			return
		}

		if t, ok := st.buckets[bucket]; ok {
			for _, it := range items {
				if old, replaced := t.ReplaceOrInsert(it); replaced {
					st.snapGarbage += st.recordSize(bucket, old, st.snapTx)
				}
			}
		} else {
			st.buckets[bucket] = buildTree(items)
		}

		items = items[:0]
	}

	// a compaction writes the compacted file at once, thus valid records never follow a bad one
	end, _, err := scanFile(snapF, snapSize, false, func(rec record, off int64, raw []byte) error {
		if rec.kind != setKeyValue {
			return fmt.Errorf("unexpected record kind %d in compacted file %s", rec.kind, db.snapPath)
		}

		if rec.tx < st.snapTx {
			return fmt.Errorf("tx must be monotonic in %s", db.snapPath)
		}

		st.snapTx = rec.tx
		if !haveBucket || unsafeString(rec.bucket) != bucket {
			flush()
			bucket = string(rec.bucket)
			haveBucket = true
		}

		items = append(items, item{key: st.keys.add(unsafeString(rec.key)), off: uint64(rec.valOff), len: rec.valLen})
		return nil
	})

	if errors.Is(err, errTornRecord) {
		slog.Warn("tdb compacted file ends with an incomplete record, which is ignored", "file", db.snapPath, "offset", end, "size", snapSize)
		st.snapGarbage += snapSize - end
		err = nil
	}

	if err != nil {
		return false, err
	}

	flush()
	items = nil
	st.snap = &dataFile{f: snapF}
	st.snap.size.Store(snapSize)

	info, err = logF.Stat()
	if err != nil {
		return false, err
	}

	logSize := info.Size()
	tx := st.snapTx
	var lastTx uint64
	end, skipped, err := scanFile(logF, logSize, true, func(rec record, off int64, raw []byte) error {
		if rec.tx < lastTx {
			return fmt.Errorf("tx must be monotonic in %s", db.logPath)
		}

		lastTx = rec.tx
		if rec.tx <= st.snapTx {
			// the record is older than the compacted file, which is the case when a compaction did not replace the log
			st.logGarbage += int64(len(raw))
			return nil
		}

		st.apply(rec, int64(len(raw)), rec.valOff)
		tx = rec.tx
		return nil
	})

	if skipped > 0 {
		slog.Warn("tdb log holds records behind an incomplete one, which is skipped", "file", db.logPath, "skipped", skipped)
		st.logGarbage += skipped
	}

	if errors.Is(err, errTornRecord) {
		// a crash while appending, the next record must not follow the garbage
		slog.Warn("tdb log ends with an incomplete record, which is removed", "file", db.logPath, "offset", end, "size", logSize)
		if err = logF.Truncate(end); err == nil {
			logSize = end
		}
	}

	if err != nil {
		return false, err
	}

	st.log = &dataFile{f: logF}
	st.log.size.Store(logSize)

	db.st = st
	db.tx = tx

	entries := 0
	for _, t := range st.buckets {
		entries += t.Len()
	}

	slog.Info("tdb opened", "dir", db.dir, "entries", entries, "file-size", snapSize+logSize, "garbage", st.garbage(), "duration", time.Since(start))
	return skipped > 0, nil
}

// openLocked opens the data file at path and locks it, see lockFile. A former version may replace the file, while
// this waits for the lock, then it locks the replacement.
func openLocked(path string) (*os.File, error) {
	for {
		f, err := openDataFile(path, os.O_RDWR|os.O_CREATE)
		if err != nil {
			return nil, err
		}

		if err := lockFile(f); err != nil {
			_ = f.Close()
			return nil, err
		}

		locked, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return nil, err
		}

		current, err := os.Stat(path)
		if err == nil && os.SameFile(locked, current) {
			return f, nil
		}

		_ = f.Close()
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
}

// recoverFiles completes or rolls back a compaction, which a crash has interrupted while it replaced a file. A
// compaction replaces the compacted file first and the log afterwards, and either one with a complete file, see
// replaceFile.
func recoverFiles(dir string) error {
	for _, name := range []string{snapName, logName} {
		path := filepath.Join(dir, name)
		next := path + nextSuffix
		old := path + oldSuffix
		if _, err := os.Stat(path); err == nil {
			// the compaction has not replaced the file or has already removed the former one
			for _, p := range []string{next, old} {
				if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
					return err
				}
			}

			continue
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		switch {
		case exists(next):
			slog.Warn("tdb completes an interrupted compaction", "file", path)
			if err := os.Rename(next, path); err != nil {
				return err
			}

			if err := os.Remove(old); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		case exists(old):
			slog.Warn("tdb rolls back an interrupted compaction", "file", path)
			if err := os.Rename(old, path); err != nil {
				return err
			}
		}
	}

	return nil
}

// removeTmpFiles removes the temporary files of compactions of former versions.
func removeTmpFiles(dir string) error {
	tmpFiles, err := filepath.Glob(filepath.Join(dir, "*.compact.tmp"))
	if err != nil {
		return err
	}

	for _, f := range tmpFiles {
		if err := os.Remove(f); err != nil {
			return err
		}
	}

	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Sync causes to fsync the write ahead log. This probably hurts performance and is likely unneeded, depending
// on what you need. If you need maximum security, you likely want to sync after each write access.
// Normally, you can just trust on pwrite and that the ups, fs and os do reasonable things - at worst, you
// have a backup anyway. The writes, which have returned before, stay durable, when a compaction replaces the log
// afterwards, see install.
func (db *DB) Sync() error {
	db.mu.RLock()
	log := db.st.log
	size := log.size.Load()
	db.syncing.Add(1)
	db.mu.RUnlock()

	err := log.f.Sync()
	if err == nil {
		storeMax(&log.synced, size)
	}

	db.syncing.Add(-1)
	if err != nil {
		return err
	}

	return db.syncRenames()
}

// syncRenames syncs the directory, if install has replaced the log since the last time.
func (db *DB) syncRenames() error {
	db.syncedDirMu.Lock()
	defer db.syncedDirMu.Unlock()

	renames := db.renames.Load()
	if db.syncedDir.Load() >= renames {
		return nil
	}

	if err := syncDir(db.dir); err != nil {
		return err
	}

	db.syncedDir.Store(renames)
	return nil
}

// storeMax stores n, if it is larger than the value.
func storeMax(v *atomic.Int64, n int64) {
	for {
		cur := v.Load()
		if cur >= n || v.CompareAndSwap(cur, n) {
			return
		}
	}
}

func (db *DB) writable() error {
	if db.closed {
		return ErrClosed
	}

	return db.broken
}

// Set appends the entry to the log and updates the index. Setting the stored value again is a no-op.
func (db *DB) Set(bucket, key string, value []byte) error {
	if size := recordSize(^uint64(0), len(bucket), len(key), uint32(len(value))) - headerSize; size > maxPayloadSize || len(value) > maxPayloadSize {
		return fmt.Errorf("tdb entry too large: %d > %d", size, maxPayloadSize)
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.writable(); err != nil {
		return err
	}

	st := db.st
	t, ok := st.buckets[bucket]
	var old item
	var found bool
	if ok {
		old, found = t.Get(item{key: key})
	}

	if found && int(old.len) == len(value) && len(value) <= maxCompareSize {
		equal, err := db.equal(st.value(old), value)
		if err != nil {
			return err
		}

		if equal {
			return nil
		}
	}

	tx := db.tx + 1
	var valOff int
	db.wbuf, valOff = appendRecord(db.wbuf[:0], setKeyValue, tx, bucket, key, value)
	off, err := db.append(st, db.wbuf)
	if cap(db.wbuf) > 1024*1024 {
		db.wbuf = nil
	}

	if err != nil {
		return err
	}

	db.tx = tx
	if !ok {
		t = newTree()
		st.buckets[strings.Clone(bucket)] = t
	}

	k := old.key
	if found {
		st.drop(bucket, old, tx)
	} else {
		k = st.keys.add(key)
	}

	t.ReplaceOrInsert(item{key: k, off: inLog | uint64(off+int64(valOff)), len: uint32(len(value))})
	db.maybeCompact()
	return nil
}

// equal compares the stored value with the given one, the caller holds the write lock.
func (db *DB) equal(v Value, value []byte) (bool, error) {
	if db.cbuf == nil {
		db.cbuf = make([]byte, 32*1024)
	}

	for pos := 0; pos < len(value); {
		n := min(len(db.cbuf), len(value)-pos)
		if _, err := v.f.ReadAt(db.cbuf[:n], v.offset+int64(pos)); err != nil {
			return false, err
		}

		if !bytes.Equal(db.cbuf[:n], value[pos:pos+n]) {
			return false, nil
		}

		pos += n
	}

	return true, nil
}

// append appends the record to the log and returns its offset, the caller holds the write lock.
func (db *DB) append(st *state, rec []byte) (int64, error) {
	off := st.log.size.Load()
	if _, err := st.log.f.WriteAt(rec, off); err != nil {
		// remove a partial record, otherwise the next one would follow garbage
		if terr := st.log.f.Truncate(off); terr != nil {
			db.broken = fmt.Errorf("tdb log is broken, a write failed and truncating it failed as well: %w", errors.Join(err, terr))
			slog.Error("tdb write failed and unable to truncate the log", "file", db.logPath, "err", db.broken)
		}

		return 0, err
	}

	st.log.size.Store(off + int64(len(rec)))
	return off, nil
}

// Stat returns the Value metadata for the given key without reading the payload.
// The returned Value carries the exact byte length via Len().
func (db *DB) Stat(bucket, key string) (Value, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	st := db.st
	t, ok := st.buckets[bucket]
	if !ok {
		return Value{}, false
	}

	it, ok := t.Get(item{key: key})
	if !ok {
		return Value{}, false
	}

	return st.value(it), true
}

func (db *DB) Exists(bucket, key string) bool {
	_, ok := db.Stat(bucket, key)
	return ok
}

func (db *DB) Get(bucket, key string) std.Option[io.ReadCloser] {
	val, ok := db.Stat(bucket, key)
	if !ok {
		return std.Option[io.ReadCloser]{}
	}

	return std.Some(val.NewReader())
}

// Buckets returns the names of the buckets in ascending order.
func (db *DB) Buckets() iter.Seq[string] {
	return func(yield func(string) bool) {
		db.mu.RLock()
		names := make([]string, 0, len(db.st.buckets))
		for name := range db.st.buckets {
			names = append(names, name)
		}
		db.mu.RUnlock()

		slices.Sort(names)
		for _, name := range names {
			if !yield(name) {
				return
			}
		}
	}
}

func (db *DB) All(bucket string) iter.Seq[IndexEntry] {
	return db.Ascend(bucket)
}

func (db *DB) Ascend(bucket string) iter.Seq[IndexEntry] {
	st, snapshot, ok := db.snapshot(bucket)
	if !ok {
		return func(yield func(IndexEntry) bool) {}
	}

	return func(yield func(IndexEntry) bool) {
		snapshot.Ascend(entries(st, yield))
	}
}

func (db *DB) Descend(bucket string) iter.Seq[IndexEntry] {
	st, snapshot, ok := db.snapshot(bucket)
	if !ok {
		return func(yield func(IndexEntry) bool) {}
	}

	return func(yield func(IndexEntry) bool) {
		snapshot.Descend(entries(st, yield))
	}
}

func (db *DB) AscendRange(bucket, minKey, maxKey string) iter.Seq[IndexEntry] {
	if minKey == "" && maxKey == "" {
		return db.Ascend(bucket)
	}

	st, snapshot, ok := db.snapshot(bucket)
	if !ok {
		return func(yield func(IndexEntry) bool) {}
	}

	return func(yield func(IndexEntry) bool) {
		snapshot.AscendRange(item{key: minKey}, item{key: maxKey}, entries(st, yield))
	}
}

func (db *DB) DescendRange(bucket, minKey, maxKey string) iter.Seq[IndexEntry] {
	if minKey == "" && maxKey == "" {
		return db.Descend(bucket)
	}

	st, snapshot, ok := db.snapshot(bucket)
	if !ok {
		return func(yield func(IndexEntry) bool) {}
	}

	return func(yield func(IndexEntry) bool) {
		snapshot.DescendRange(item{key: maxKey}, item{key: minKey}, entries(st, yield))
	}
}

func entries(st *state, yield func(IndexEntry) bool) func(item) bool {
	return func(it item) bool {
		return yield(IndexEntry{key: it.key, val: st.value(it)})
	}
}

func (db *DB) Len(bucket string) int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	t, ok := db.st.buckets[bucket]
	if !ok {
		return 0
	}

	return t.Len()
}

// Delete removes the entry from the in-memory index and adds that to the WAL. Deleting non-existing entries is a no-op.
func (db *DB) Delete(bucket, key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.writable(); err != nil {
		return err
	}

	st := db.st
	t, ok := st.buckets[bucket]
	if !ok {
		return nil
	}

	old, found := t.Get(item{key: key})
	if !found {
		return nil
	}

	tx := db.tx + 1
	db.wbuf, _ = appendRecord(db.wbuf[:0], removeKeyValue, tx, bucket, key, nil)
	if _, err := db.append(st, db.wbuf); err != nil {
		return err
	}

	db.tx = tx
	t.Delete(old)
	st.drop(bucket, old, tx)
	st.removal(int64(len(db.wbuf)))
	db.maybeCompact()
	return nil
}

// Stats returns the sizes of the files and the state of the compaction.
func (db *DB) Stats() Stats {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return Stats{
		FileSize:       db.st.snap.size.Load() + db.st.log.size.Load(),
		Garbage:        db.st.garbage(),
		Compactions:    db.compactions.Load(),
		LogCompactions: db.logCompactions.Load(),
		Compacting:     db.compacting.Load(),
	}
}

// snapshot returns a lazy copy-on-write clone of the bucket tree together with its state, which can be iterated
// without further locking. Clone mutates the copy-on-write context of the original tree, thus it must be exclusive
// to all writers and other clones, see also [btree.BTreeG.Clone].
func (db *DB) snapshot(bucket string) (*state, *btree.BTreeG[item], bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	st := db.st
	t, ok := st.buckets[bucket]
	if !ok {
		return nil, nil, false
	}

	return st, t.Clone(), true
}

// Close waits for a running compaction to stop and closes the files. Readers of values fail afterwards.
func (db *DB) Close() error {
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return nil
	}

	db.closed = true
	db.mu.Unlock()

	db.closing.Store(true)
	db.bg.Wait()
	// an explicit Compact sees the closed DB as well
	db.compactMu.Lock()
	db.compactMu.Unlock()

	db.mu.Lock()
	st := db.st
	db.mu.Unlock()

	// close every file, even if one of them fails, and sync the log before
	var errs []error
	if err := st.log.f.Sync(); err != nil {
		errs = append(errs, fmt.Errorf("cannot sync log: %w", err))
	}

	if err := db.syncRenames(); err != nil {
		errs = append(errs, fmt.Errorf("cannot sync directory: %w", err))
	}

	if err := st.log.f.Close(); err != nil {
		errs = append(errs, fmt.Errorf("cannot close log: %w", err))
	}

	if err := st.snap.f.Close(); err != nil {
		errs = append(errs, fmt.Errorf("cannot close compacted file: %w", err))
	}

	// release the directory only now, otherwise another Open could succeed before our files are closed
	if err := db.dirLock.Close(); err != nil {
		errs = append(errs, fmt.Errorf("cannot release directory lock: %w", err))
	}

	db.procLock.Unlock()

	return errors.Join(errs...)
}
