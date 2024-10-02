package tdb

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/google/btree"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xmaps"
	"io"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var globalDirLocks = xmaps.NewConcurrentMap[string, *sync.Mutex]()

type IndexEntry struct {
	key string
	val Value
}

type DB struct {
	buckets           *xmaps.ConcurrentMap[string, *btree.BTreeG[IndexEntry]]
	wal               *WAL
	compacted         *WAL
	compactFile       string
	walFile           string
	strDedupTable     *xmaps.ConcurrentMap[string, string]
	tx                atomic.Uint64
	btreeSnapshotLock sync.RWMutex
	compactLock       sync.Mutex
	dir               string
}

func Open(dir string) (*DB, error) {
	if err := os.MkdirAll(dir, 0777); err != nil && !os.IsExist(err) {
		return nil, err
	}

	mutex, _ := globalDirLocks.LoadOrStore(dir, &sync.Mutex{})
	if ok := mutex.TryLock(); !ok {
		return nil, fmt.Errorf("tdb directory is already opened by another instance: %s", dir)
	}

	db := &DB{
		buckets:       xmaps.NewConcurrentMap[string, *btree.BTreeG[IndexEntry]](),
		strDedupTable: xmaps.NewConcurrentMap[string, string](),
		dir:           dir,
		walFile:       filepath.Join(dir, "tdb.wal"),
		compactFile:   filepath.Join(dir, "data.tdb"),
	}

	// initialize from compacted file
	var lastTx uint64
	comfdb, err := OpenWAL(db.compactFile, func(entry *Node) {
		tx := entry.tx
		lastTx = tx // guaranteed to be monotonic
		bucket := db.strOf(entry.bucket)
		key := db.strOf(entry.key)

		if entry.kind != setKeyValue {
			panic(fmt.Errorf("unexpected node kind in compacted log file"))
		}

		tree, ok := db.buckets.Load(bucket)
		if !ok {
			tree = newBtree()
			db.buckets.Store(bucket, tree)
		}

		tree.ReplaceOrInsert(IndexEntry{
			key: key,
			val: entry.Value(),
		})
	})
	if err != nil {
		return nil, err
	}

	db.compacted = comfdb

	// now work through the actual WAL
	var actualTx uint64
	waldb, err := OpenWAL(db.walFile, func(entry *Node) {
		if entry.tx <= lastTx {
			// do not apply values, which are older than the compacted db
			return
		}

		bucket := db.strOf(entry.bucket)
		key := db.strOf(entry.key)

		tree, ok := db.buckets.Load(bucket)
		if !ok {
			tree = newBtree()
			db.buckets.Store(bucket, tree)
		}

		if entry.kind == setKeyValue {
			tree.ReplaceOrInsert(IndexEntry{
				key: key,
				val: entry.Value(),
			})
		} else {
			tree.Delete(IndexEntry{
				key: key,
			})
		}

		actualTx = entry.tx
	})

	if err != nil {
		return nil, err
	}

	db.wal = waldb

	db.tx.Store(actualTx)

	requiresCompaction := false
	if info, err := os.Stat(db.walFile); err == nil {
		requiresCompaction = info.Size() > 1024*1024*128 // 128mib
	}

	if requiresCompaction {
		slog.Info("tdb WAL reached threshold, starting compaction", "path", db.dir)
		start := time.Now()
		if err := db.Compact(); err != nil {
			return nil, err
		}

		slog.Info("tdb compaction complete", "path", db.dir, "duration", time.Since(start))
	}

	return db, nil
}

// Sync causes to fsync the write ahead log. This probably hurts performance and is likely unneeded, depending
// on what you need. If you need maximum security, you likely want to sync after each write access.
// Normally, you can just trust on pwrite and that the ups, fs and os do reasonable things - at worst, you
// have a backup anyway.
func (db *DB) Sync() error {
	return db.wal.f.Sync()
}

// Set writes to the WAL and updates in-memory index tree.
func (db *DB) Set(bucket, key string, value []byte) error {
	// lock on our trees
	db.btreeSnapshotLock.Lock()
	defer db.btreeSnapshotLock.Unlock()

	// TODO it may be clever to compare the value before writing, so that writing the same data is a no-op

	bucketTree, ok := db.buckets.Load(bucket)
	if !ok {
		bucketTree, _ = db.buckets.LoadOrStore(bucket, newBtree())
	}

	oldEntry, ok := bucketTree.Get(IndexEntry{key: key})
	// a little micro-optimization for small buffers, which we can stack-allocate
	const maxCheck = 1024 * 128 // this could be at most 10mib, before going to heap, however, if the compiler fails to inline it escapes anyway, but Copy code is all non-virtual and deadly simple
	if ok && oldEntry.val.Len() == len(value) && len(value) < maxCheck {
		var tmp [maxCheck]byte
		slice := tmp[:len(value)]
		if err := oldEntry.val.Copy(slice); err != nil {
			return err
		}

		if bytes.Equal(slice, value) {
			// short circuit, nothing changed
			return nil
		}
	}

	tx := db.tx.Add(1)
	val, err := db.wal.SetWithTx(unsafeStr(bucket), unsafeStr(key), value, tx)
	if err != nil {
		return err
	}

	bucketTree.ReplaceOrInsert(IndexEntry{
		key: key,
		val: val,
	})

	return nil
}

func (db *DB) Exists(bucket, key string) bool {
	// lock on our trees, but read-only.
	db.btreeSnapshotLock.RLock()
	defer db.btreeSnapshotLock.RUnlock()

	tree, ok := db.buckets.Load(bucket)
	if !ok {
		return false
	}

	_, ok = tree.Get(IndexEntry{
		key: key,
	})

	return ok
}

func (db *DB) Get(bucket, key string) std.Option[io.ReadCloser] {
	// lock on our trees, but read-only.
	db.btreeSnapshotLock.RLock()
	defer db.btreeSnapshotLock.RUnlock()

	tree, ok := db.buckets.Load(bucket)
	if !ok {
		return std.Option[io.ReadCloser]{}
	}

	entry, ok := tree.Get(IndexEntry{
		key: key,
	})

	if !ok {
		return std.Option[io.ReadCloser]{}
	}

	return std.Some(entry.val.NewReader())
}

func (db *DB) Buckets() iter.Seq[string] {
	return func(yield func(string) bool) {
		for k, _ := range db.buckets.All() {
			if !yield(k) {
				return
			}
		}
	}
}

func (db *DB) All(bucket string) iter.Seq[IndexEntry] {
	db.btreeSnapshotLock.RLock()
	tree, ok := db.buckets.Load(bucket)
	db.btreeSnapshotLock.RUnlock()

	if !ok {
		return func(yield func(IndexEntry) bool) {}
	}

	snapshot := tree.Clone()

	return func(yield func(IndexEntry) bool) {
		snapshot.Ascend(yield)
	}
}

func (db *DB) Range(bucket, minKey, maxKey string) iter.Seq[IndexEntry] {
	if minKey == "" && maxKey == "" {
		return db.All(bucket)
	}

	db.btreeSnapshotLock.RLock()
	tree, ok := db.buckets.Load(bucket)
	db.btreeSnapshotLock.RUnlock()

	if !ok {
		return func(yield func(IndexEntry) bool) {}
	}

	snapshot := tree.Clone()

	return func(yield func(IndexEntry) bool) {
		snapshot.AscendRange(IndexEntry{key: minKey}, IndexEntry{key: maxKey}, yield)
	}
}

// Delete removes the entry from the in-memory index and adds that to the WAL. Deleting non-existing entries is a no-op.
func (db *DB) Delete(bucket, key string) error {
	db.btreeSnapshotLock.Lock()
	defer db.btreeSnapshotLock.Unlock()

	// avoid I/O if we should delete stuff, which is not in-memory anyway
	tree, ok := db.buckets.Load(bucket)
	if !ok {
		return nil
	}

	_, removed := tree.Delete(IndexEntry{
		key: key,
	})

	if !removed {
		return nil
	}

	return db.wal.DeleteWithTx(unsafeStr(bucket), unsafeStr(key), db.tx.Add(1))
}

// Compact writes a snapshot of the current in-memory state and persistent values to a new disk file.
// It mostly runs concurrently, however if changes are made during compaction, the WAL will not be removed afterwards.
// The DB performs also compaction on startup, which has the following advantages:
//   - there is no hard deadline, like on shutdown
//   - there are no concurrent writes
func (db *DB) Compact() error {
	// we don't want concurrent compaction runs
	db.compactLock.Lock()
	defer db.compactLock.Unlock()

	var tmp [16]byte
	if _, err := rand.Read(tmp[:]); err != nil {
		return err
	}

	tmpFile := filepath.Join(db.dir, hex.EncodeToString(tmp[:])+".compact.tmp")
	compactWal, err := OpenWAL(tmpFile, nil)
	if err != nil {
		return fmt.Errorf("cannot open WAL for compacted snapshot: %w", err)
	}

	trees, tx := db.snapshotTrees()

	var valueTmp bytes.Buffer
	var copyBuf [32 * 1024]byte
	for bucketName, bucketTree := range trees {
		bucketTree.Ascend(func(item IndexEntry) bool {
			valueTmp.Reset()
			r := item.val.NewReader()
			_, e := io.CopyBuffer(&valueTmp, r, copyBuf[:])
			if e != nil {
				err = e
				return false
			}

			if _, e := compactWal.SetWithTx(unsafeStr(bucketName), unsafeStr(item.key), valueTmp.Bytes(), tx); e != nil {
				err = e
				return false
			}

			return true
		})

	}

	if err != nil {
		return fmt.Errorf("cannot compact TDB: %w", err)
	}

	if err := compactWal.Close(); err != nil {
		return fmt.Errorf("cannot close compact WAL: %w", err)
	}

	// this one is now written new, replace it
	_ = db.compacted.Close()

	if err := os.Rename(tmpFile, db.compactFile); err != nil {
		return fmt.Errorf("cannot atomic rename TDB compaction file: %w", err)
	}

	// actually we should also fsync the directory here to persist the inode file content after the rename, but probably that is nitpicking in our context

	// now, lets take a look, if we just can delete the WAL
	db.btreeSnapshotLock.Lock()
	defer db.btreeSnapshotLock.Unlock()

	if db.tx.Load() == tx {
		// nothing changed, thus no other appends have happened. we just can clean up the wal
		_ = db.wal.Close()
		if err := os.Remove(db.walFile); err != nil {
			return fmt.Errorf("cannot remove WAL file: %w", err)
		}

	}

	// otherwise the old WAL obviously was appended, thus don't touch it. Next time we are started, older tx entries are ignored during replay

	// either compact file or also the wal file changed, thus our index value offsets are now illegal, we need to reload everything again
	db2, err := Open(db.dir)
	if err != nil {
		return fmt.Errorf("cannot re-open TDB: %w", err)
	}

	db.compacted = db2.compacted
	db.wal = db2.wal
	db.buckets = db2.buckets
	db.strDedupTable = db2.strDedupTable

	return err
}

func (db *DB) snapshotTrees() (map[string]*btree.BTreeG[IndexEntry], uint64) {
	db.btreeSnapshotLock.Lock()
	defer db.btreeSnapshotLock.Unlock()

	res := map[string]*btree.BTreeG[IndexEntry]{}
	for bucketName, bucketTree := range db.buckets.All() {
		res[bucketName] = bucketTree.Clone()
	}

	return res, db.tx.Load()
}

func (db *DB) strOf(buf []byte) string {
	str, ok := db.strDedupTable.Load(unsafe.String(&buf[0], len(buf)))
	if ok {
		return str
	}

	str = string(buf)
	db.strDedupTable.Store(str, str)
	return str
}

func (db *DB) Close() error {
	db.btreeSnapshotLock.Lock()
	defer db.btreeSnapshotLock.Unlock()

	if lock, ok := globalDirLocks.Load(db.dir); ok {
		lock.Unlock()
	}

	if err := db.Sync(); err != nil {
		return err
	}

	if err := db.wal.Close(); err != nil {
		return fmt.Errorf("cannot close WAL: %w", err)
	}

	if err := db.compacted.Close(); err != nil {
		return fmt.Errorf("cannot close compacted: %w", err)
	}

	return nil
}

func unsafeStr(str string) []byte {
	d := unsafe.StringData(str)
	return unsafe.Slice(d, len(str))
}

func newBtree() *btree.BTreeG[IndexEntry] {
	return btree.NewG[IndexEntry](2, func(a, b IndexEntry) bool {
		return strings.Compare(a.key, b.key) < 0
	})
}
