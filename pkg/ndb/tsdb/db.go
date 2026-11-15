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
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rogpeppe/go-internal/lockedfile"

	"go.wdy.de/nago/pkg/ndb"
)

// DB is a tsdb engine instance: a directory of buckets, each holding typed
// columns. It is exclusively locked to a single process via a lock file and
// safe for concurrent use.
type DB struct {
	dir      string
	opts     Options
	fs       *ndb.FilePool
	lockFile *lockedfile.File

	// columns holds a slot for every opened column keyed by "bucket/column". It is
	// a sync.Map so that the overwhelmingly common cache-hit path (a stateless
	// caller re-resolving a column on every operation) is lock-free and does not
	// suffer the reader cache-line contention of a RWMutex under concurrency.
	columns sync.Map // map[string]*columnSlot
	// openMu is held for reading while a column is opened, so that columns open side by side, and for writing by
	// Close and the drops, which wait for the opens running.
	openMu sync.RWMutex
	// closed is checked lock-free on the hit path.
	closed atomic.Bool

	compactor *compactor
	idleStop  chan struct{}
	idleDone  chan struct{}
}

// columnSlot is the place of one column in DB.columns. The first caller opens the column, the others of the same name
// wait for ready; callers of other names do not wait.
type columnSlot struct {
	ready chan struct{}
	col   atomic.Pointer[Column]
	err   error
}

// errNoColumn is the outcome of an open without create of a column that is not on disk.
var errNoColumn = errors.New("tsdb: no such column")

// Open opens or creates a tsdb instance rooted at dir.
func Open(dir string, opts Options) (*DB, error) {
	opts.resolve()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("tsdb: create dir: %w", err)
	}
	lockFile, err := lockedfile.Create(filepath.Join(dir, ".lock"))
	if err != nil {
		return nil, fmt.Errorf("tsdb: store already locked by another process: %w", err)
	}
	db := &DB{
		dir:      dir,
		opts:     opts,
		fs:       opts.FilePool,
		lockFile: lockFile,
	}
	db.compactor = newCompactor(db)
	if opts.IdleAfter > 0 {
		db.idleStop, db.idleDone = make(chan struct{}), make(chan struct{})
		go db.idleLoop()
	}
	return db, nil
}

func columnKey(bucket, column string) string { return bucket + "/" + column }

// resolve returns the open column of key, opening it with open if it is not open yet. Concurrent calls for the same
// key wait for a single open; calls for other keys open theirs at the same time. A failed open leaves no slot, so
// that the next call tries again.
func (db *DB) resolve(key string, open func() (*Column, error)) (*Column, error) {
	for {
		if v, ok := db.columns.Load(key); ok {
			slot := v.(*columnSlot)
			if c := slot.col.Load(); c != nil {
				return c, nil
			}
			<-slot.ready
			if c := slot.col.Load(); c != nil {
				return c, nil
			}
			continue // the open failed; the next round opens again
		}
		db.openMu.RLock()
		if db.closed.Load() {
			db.openMu.RUnlock()
			return nil, errClosed
		}
		slot := &columnSlot{ready: make(chan struct{})}
		if _, loaded := db.columns.LoadOrStore(key, slot); loaded {
			db.openMu.RUnlock()
			continue
		}
		c, err := open()
		if err != nil {
			slot.err = err
			db.columns.Delete(key)
		} else {
			slot.col.Store(c)
		}
		close(slot.ready)
		db.openMu.RUnlock()
		return c, err
	}
}

// Column opens (creating if necessary) the column with the given scheme in the
// given bucket. If the column already exists its recorded schema must match the
// requested scheme and decimals, otherwise an error is returned.
//
// The already-open case is a lock-free map lookup plus a schema-compatibility
// check; name and schema validation only run when a column must actually be
// opened, so re-resolving a live column on every operation is cheap even under
// heavy concurrency. Different columns open at the same time; the same column is opened once, and every call gets the
// same *Column until it is dropped or the DB is closed.
func (db *DB) Column(bucket, column string, schema Schema) (*Column, error) {
	if db.closed.Load() {
		return nil, errClosed
	}
	key := columnKey(bucket, column)
	// fast path: already open — lock-free load, then compatibility check.
	if v, ok := db.columns.Load(key); ok {
		if c := v.(*columnSlot).col.Load(); c != nil {
			return checkSchema(c, schema)
		}
	}

	if err := validateName(bucket); err != nil {
		return nil, err
	}
	if err := validateName(column); err != nil {
		return nil, err
	}
	if err := schema.validate(); err != nil {
		return nil, err
	}
	schema.Version = schemaVersion
	c, err := db.resolve(key, func() (*Column, error) { return db.openColumnDir(bucket, column, &schema) })
	if err != nil {
		return nil, err
	}
	return checkSchema(c, schema)
}

func checkSchema(c *Column, schema Schema) (*Column, error) {
	if c.schema.Scheme != schema.Scheme || c.schema.Decimals != schema.Decimals {
		return nil, errColumnExists
	}
	return c, nil
}

// openColumnDir opens the column from its directory. With create, a column, which is not on disk, is created with
// it; without, the open fails with errNoColumn then. The open reads schema.json and the end of the open chunk.
func (db *DB) openColumnDir(bucket, column string, create *Schema) (*Column, error) {
	dir := filepath.Join(db.dir, bucket, column)
	schema, ok, err := loadSchema(db.fs, dir)
	if err != nil {
		return nil, err
	}
	switch {
	case ok:
		if create != nil && (schema.Scheme != create.Scheme || schema.Decimals != create.Decimals) {
			return nil, errColumnExists
		}
	case create == nil:
		return nil, errNoColumn
	default:
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("tsdb: create column dir: %w", err)
		}
		if err := storeSchema(db.fs, dir, *create); err != nil {
			return nil, err
		}
		if err := db.fs.WriteFile(filepath.Join(dir, headWALName), nil); err != nil {
			return nil, fmt.Errorf("tsdb: create head wal: %w", err)
		}
		schema = *create
	}
	return openColumn(db, bucket, column, dir, schema)
}

// LookupColumn returns an already-open or on-disk column without creating a new
// one. ok is false if the column does not exist.
func (db *DB) LookupColumn(bucket, column string) (*Column, bool, error) {
	if db.closed.Load() {
		return nil, false, errClosed
	}

	key := columnKey(bucket, column)
	// fast path: already open — lock-free.
	if v, ok := db.columns.Load(key); ok {
		if c := v.(*columnSlot).col.Load(); c != nil {
			return c, true, nil
		}
	}

	if err := validateName(bucket); err != nil {
		return nil, false, err
	}
	if err := validateName(column); err != nil {
		return nil, false, err
	}
	c, err := db.resolve(key, func() (*Column, error) { return db.openColumnDir(bucket, column, nil) })
	if errors.Is(err, errNoColumn) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return c, true, nil
}

// SeriesColumns lists all columns present on disk as "bucket/column".
func (db *DB) SeriesColumns() ([]string, error) {
	entries, err := os.ReadDir(db.dir)
	if err != nil {
		return nil, fmt.Errorf("tsdb: list buckets: %w", err)
	}
	var out []string
	for _, be := range entries {
		if !be.IsDir() || strings.HasPrefix(be.Name(), ".") {
			continue
		}
		bucket := be.Name()
		cols, err := os.ReadDir(filepath.Join(db.dir, bucket))
		if err != nil {
			return nil, fmt.Errorf("tsdb: list columns: %w", err)
		}
		for _, ce := range cols {
			if !ce.IsDir() {
				continue
			}
			if _, ok, _ := loadSchema(db.fs, filepath.Join(db.dir, bucket, ce.Name())); ok {
				out = append(out, bucket+"/"+ce.Name())
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

// DropColumn removes the column from the disk and from the open columns. A later
// call of Column creates it empty again, and a handle of the dropped column only
// returns errors. A column, which does not exist, is no error.
func (db *DB) DropColumn(bucket, column string) error {
	if err := validateName(bucket); err != nil {
		return err
	}
	if err := validateName(column); err != nil {
		return err
	}
	db.openMu.Lock()
	defer db.openMu.Unlock()
	if db.closed.Load() {
		return errClosed
	}
	err := db.forget(columnKey(bucket, column))
	if rmErr := db.removeColumnDir(filepath.Join(db.dir, bucket, column)); rmErr != nil {
		return errors.Join(err, fmt.Errorf("tsdb: remove column: %w", rmErr))
	}
	// only an empty bucket goes with its last column; any other error says the same
	_ = os.Remove(filepath.Join(db.dir, bucket))
	return err
}

// DropBucket removes the bucket with all its columns from the disk and from the
// open columns. See [DB.DropColumn].
func (db *DB) DropBucket(bucket string) error {
	if err := validateName(bucket); err != nil {
		return err
	}
	db.openMu.Lock()
	defer db.openMu.Unlock()
	if db.closed.Load() {
		return errClosed
	}
	var err error
	db.columns.Range(func(k, v any) bool {
		if c := v.(*columnSlot).col.Load(); c != nil && c.bucket == bucket {
			err = errors.Join(err, db.forget(k.(string)))
		}
		return true
	})
	dir := filepath.Join(db.dir, bucket)
	cols, rmErr := os.ReadDir(dir)
	if os.IsNotExist(rmErr) {
		rmErr = nil
	}
	if rmErr != nil {
		return errors.Join(err, fmt.Errorf("tsdb: list bucket: %w", rmErr))
	}
	for _, ce := range cols {
		if ce.IsDir() {
			rmErr = errors.Join(rmErr, db.removeColumnDir(filepath.Join(dir, ce.Name())))
		}
	}
	if rmErr = errors.Join(rmErr, os.RemoveAll(dir)); rmErr != nil {
		return errors.Join(err, fmt.Errorf("tsdb: remove bucket: %w", rmErr))
	}
	return err
}

// removeColumnDir removes the files of a column through the pool, then its directory.
func (db *DB) removeColumnDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			err = errors.Join(err, db.fs.Remove(filepath.Join(dir, e.Name())))
		}
	}
	return errors.Join(err, os.RemoveAll(dir))
}

// forget removes an open column from the map and makes its handle fail. The caller
// holds openMu, so that no column of that name is opened meanwhile.
func (db *DB) forget(key string) error {
	v, ok := db.columns.LoadAndDelete(key)
	if !ok {
		return nil
	}
	if c := v.(*columnSlot).col.Load(); c != nil {
		return c.markDropped()
	}
	return nil
}

// openColumns returns the open columns.
func (db *DB) openColumns() []*Column {
	var cols []*Column
	db.columns.Range(func(_, v any) bool {
		if c := v.(*columnSlot).col.Load(); c != nil {
			cols = append(cols, c)
		}
		return true
	})
	return cols
}

// Flush forces a synchronous flush + compaction of the given column, folding
// its head into sealed chunks and reclaiming space. Unlike Sync, DB.Close and the
// background compaction, it also finalizes the open pending chunk, so that the
// next point begins a new chunk file. Useful right after a known bulk rewrite.
func (c *Column) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.flushLocked()
}

// Sync saves the points of the open block, which only the memory holds, as one continuation record behind the open
// chunk and frees their memory. It neither seals the block nor finalizes the chunk: the points after it continue the
// open block, which is sealed when it is full, as if no Sync had run, so it can be called in a fixed rhythm without
// adding a block or a file each time. A Sync writes into one file of the column, and with nothing in memory it writes
// nothing. After a crash of the process, the next open recovers every point written before a Sync and continues the
// open chunk; a crash during a Sync costs only the points since the Sync before. No fsync is done (see the durability
// model in the README). The split policy is asked at the first point after a Sync. Out-of-order points are not
// affected: they are in the write-ahead log of the head already.
func (c *Column) Sync() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dropped {
		return errColumnGone
	}
	return c.app.sync()
}

// ReleaseIdle saves every open column, into which nothing was written for Options.IdleAfter by the clock
// Options.Now, like Column.Sync, and frees the memory of its unsaved points. With IdleAfter 0 (off) it does nothing.
// A read does not keep a column from being idle, and a held *Column stays valid. With IdleAfter set, the DB
// calls it every IdleAfter by itself.
func (db *DB) ReleaseIdle() error {
	if db.closed.Load() {
		return errClosed
	}
	idle := db.opts.IdleAfter
	if idle <= 0 {
		return nil
	}
	now := db.opts.Now().UnixNano()
	var err error
	for _, c := range db.openColumns() {
		if time.Duration(now-c.lastWrite.Load()) < idle {
			continue
		}
		c.mu.Lock()
		if !c.dropped {
			err = errors.Join(err, c.app.sync())
		}
		c.mu.Unlock()
	}
	return err
}

func (db *DB) idleLoop() {
	defer close(db.idleDone)
	t := time.NewTicker(db.opts.IdleAfter)
	defer t.Stop()
	for {
		select {
		case <-db.idleStop:
			return
		case <-t.C:
			if err := db.ReleaseIdle(); err != nil && !errors.Is(err, errClosed) {
				slog.Error("tsdb: release idle columns failed", "err", err)
			}
		}
	}
}

// Close seals the open block of every column into its open pending chunk,
// folds the head of every column into chunks, stops the compactor, and releases
// the lock. It does not finalize a pending chunk: the next Open continues it, so
// that only the split policy ends a chunk and a restart does not begin a new file
// per column. Only a head reaching into the pending chunk rewrites it, and the
// newest of the rewritten chunks stays open. The columns do this side by side, and
// the compactions of all of them commit together, so that Close waits for their
// syncs once (runCompactions). The shared FilePool is owned by ndb and is not
// closed here.
func (db *DB) Close() error {
	if !db.closed.CompareAndSwap(false, true) {
		return nil // already closed
	}

	// wait for the opens running, so that no column is opened after the snapshot.
	db.openMu.Lock()
	cols := db.openColumns()
	db.openMu.Unlock()

	if db.idleStop != nil {
		close(db.idleStop)
		<-db.idleDone
	}
	db.compactor.stop()

	// Each column seals its open block as a block of its own and prepares the fold of a head, which was read in this
	// life; the folds of all columns then commit together, so that their syncs are bundled (runCompactions).
	var errs []error
	for _, c := range cols {
		c.mu.Lock()
	}
	prepared := make([]*compaction, len(cols))
	errs = make([]error, len(cols))
	parallel(len(cols), func(i int) {
		c := cols[i]
		if c.dropped {
			return
		}
		if errs[i] = c.app.sealBlock(); errs[i] != nil || !c.head.ready.Load() {
			return
		}
		prepared[i], errs[i] = c.prepareFold(false)
	})
	var ks []*compaction
	for _, k := range prepared {
		if k != nil {
			ks = append(ks, k)
		}
	}
	if len(ks) > 0 {
		errs = append(errs, runCompactions(db.opts.FilePool, ks))
	}
	for _, c := range cols {
		errs = append(errs, c.close())
		c.mu.Unlock()
	}

	if db.lockFile != nil {
		db.lockFile.Close()
		db.lockFile = nil
	}
	return errors.Join(errs...)
}

// validateName restricts bucket/column names to filesystem-safe identifiers.
func validateName(name string) error {
	if name == "" || len(name) > 255 {
		return errBadName
	}
	if strings.HasPrefix(name, ".") {
		return errBadName
	}
	for _, r := range name {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' ||
			r == '.' || r == '_' || r == '-' {
			continue
		}
		return errBadName
	}
	return nil
}
