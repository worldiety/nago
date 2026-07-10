// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/rogpeppe/go-internal/lockedfile"
)

// DB is a tsdb engine instance: a directory of buckets, each holding typed
// columns. It is exclusively locked to a single process via a lock file and
// safe for concurrent use.
type DB struct {
	dir      string
	opts     Options
	lockFile *lockedfile.File

	mu      sync.RWMutex
	columns map[string]*Column // key: bucket + "/" + column
	closed  bool

	compactor *compactor
}

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
		lockFile: lockFile,
		columns:  make(map[string]*Column),
	}
	db.compactor = newCompactor(db)
	return db, nil
}

func columnKey(bucket, column string) string { return bucket + "/" + column }

// Column opens (creating if necessary) the column with the given scheme in the
// given bucket. If the column already exists its recorded schema must match the
// requested scheme and decimals, otherwise an error is returned.
func (db *DB) Column(bucket, column string, schema Schema) (*Column, error) {
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

	key := columnKey(bucket, column)

	db.mu.RLock()
	if db.closed {
		db.mu.RUnlock()
		return nil, errClosed
	}
	if c, ok := db.columns[key]; ok {
		db.mu.RUnlock()
		if c.schema.Scheme != schema.Scheme || c.schema.Decimals != schema.Decimals {
			return nil, errColumnExists
		}
		return c, nil
	}
	db.mu.RUnlock()

	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed {
		return nil, errClosed
	}
	if c, ok := db.columns[key]; ok {
		if c.schema.Scheme != schema.Scheme || c.schema.Decimals != schema.Decimals {
			return nil, errColumnExists
		}
		return c, nil
	}

	dir := filepath.Join(db.dir, bucket, column)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("tsdb: create column dir: %w", err)
	}

	existing, ok, err := loadSchema(dir)
	if err != nil {
		return nil, err
	}
	if ok {
		if existing.Scheme != schema.Scheme || existing.Decimals != schema.Decimals {
			return nil, errColumnExists
		}
		schema = existing
	} else {
		if err := storeSchema(dir, schema); err != nil {
			return nil, err
		}
	}

	c, err := openColumn(db, bucket, column, dir, schema)
	if err != nil {
		return nil, err
	}
	db.columns[key] = c
	return c, nil
}

// LookupColumn returns an already-open or on-disk column without creating a new
// one. ok is false if the column does not exist.
func (db *DB) LookupColumn(bucket, column string) (*Column, bool, error) {
	if err := validateName(bucket); err != nil {
		return nil, false, err
	}
	if err := validateName(column); err != nil {
		return nil, false, err
	}
	key := columnKey(bucket, column)
	db.mu.RLock()
	if db.closed {
		db.mu.RUnlock()
		return nil, false, errClosed
	}
	if c, ok := db.columns[key]; ok {
		db.mu.RUnlock()
		return c, true, nil
	}
	db.mu.RUnlock()

	dir := filepath.Join(db.dir, bucket, column)
	schema, ok, err := loadSchema(dir)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	c, err := db.Column(bucket, column, schema)
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
			if _, ok, _ := loadSchema(filepath.Join(db.dir, bucket, ce.Name())); ok {
				out = append(out, bucket+"/"+ce.Name())
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

// Flush forces a synchronous flush + compaction of the given column, folding
// its head into sealed chunks and reclaiming space. Useful right after a known
// bulk rewrite.
func (c *Column) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.flushLocked()
}

// Close flushes all columns, stops the compactor, and releases the lock. The
// shared FilePool is owned by ndb and is not closed here.
func (db *DB) Close() error {
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return nil
	}
	db.closed = true
	cols := make([]*Column, 0, len(db.columns))
	for _, c := range db.columns {
		cols = append(cols, c)
	}
	db.mu.Unlock()

	db.compactor.stop()

	var firstErr error
	for _, c := range cols {
		c.mu.Lock()
		if err := c.flushLocked(); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := c.close(); err != nil && firstErr == nil {
			firstErr = err
		}
		c.mu.Unlock()
	}

	if db.lockFile != nil {
		db.lockFile.Close()
		db.lockFile = nil
	}
	return firstErr
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
