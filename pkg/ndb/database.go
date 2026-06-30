// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndb

import (
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/worldiety/option"
)

// engineMarker is the name of the per-engine-instance file that records which
// engine kind created the instance, so it is always reopened with the matching
// implementation.
const engineMarker = ".engine"

// DB is the central nago database: a root directory that manages many named,
// technically distinct storage engine instances ([Engine]), including several
// of the same kind. Each immediate subdirectory of the root is one engine
// instance, named after the directory.
//
// Engine instances are opened lazily on first access (see [DB.Engine]) and
// cached, so the underlying engine — and its file pool, locks, ... — is shared
// across callers and there is never more than one live instance per directory.
// The DB owns the lifecycle of every instance it opens: [DB.Close] releases
// them all, and an [Engine] cannot be closed by a consumer. After Close the DB
// handle is spent and must not be reused.
//
// DB is safe for concurrent use.
type DB struct {
	root string
	opts Options

	mu     sync.Mutex
	open   map[string]openEngine
	closed bool
}

// openEngine pairs a live engine instance with the close function its factory
// returned. The close function is retained only here so that exclusively [DB]
// controls the instance lifecycle.
type openEngine struct {
	engine Engine
	close  func() error
}

// Open inspects the given root directory and returns a DB for it. The directory
// is created if it does not exist. No engine instance is opened here; instances
// are opened lazily on first access (see [DB.Engine] / [DB.LookupEngine]).
func Open(root string, opts Options) (*DB, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("ndb: create root dir: %w", err)
	}
	return &DB{
		root: root,
		opts: opts,
		open: make(map[string]openEngine),
	}, nil
}

// Engine opens the named engine instance, creating it if it does not yet exist,
// and returns the same cached instance on subsequent calls.
//
// An existing instance is always reopened with the engine that originally
// created it (recorded in a per-instance marker file), so its on-disk format is
// never reinterpreted by a different engine. A new instance is created with the
// engine selected by [EngineOptions.Kind], falling back to [Options.DefaultKind]
// and then to the sole registered engine when unambiguous.
//
// opts.Config is passed through to the engine factory, but only when the
// instance is actually created/opened: if the instance is already open, the
// cached instance is returned and opts.Config is ignored.
func (db *DB) Engine(name string, opts EngineOptions) (Engine, error) {
	if err := validateEngineName(name); err != nil {
		return nil, err
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return nil, fmt.Errorf("ndb: database is closed")
	}

	if eng, ok := db.open[name]; ok {
		return eng.engine, nil
	}

	dir := filepath.Join(db.root, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("ndb: create engine dir: %w", err)
	}

	kind, err := db.resolveEngineKind(dir, opts.Kind)
	if err != nil {
		return nil, err
	}

	factory, resolvedKind, err := lookupEngine(kind)
	if err != nil {
		return nil, err
	}
	kind = resolvedKind

	eng, closeFn, err := factory(name, dir, opts.Config)
	if err != nil {
		return nil, fmt.Errorf("ndb: open engine %q with kind %q: %w", name, kind, err)
	}

	if err := writeEngineMarker(dir, kind); err != nil {
		if closeFn != nil {
			_ = closeFn()
		}
		return nil, err
	}

	db.open[name] = openEngine{engine: eng, close: closeFn}
	return eng, nil
}

// LookupEngine returns an already-known engine instance, opening it if it exists
// on disk. It never creates a new instance: the result is empty if no instance
// with that name exists. The recorded engine kind is used, so no options are
// required.
func (db *DB) LookupEngine(name string) (option.Opt[Engine], error) {
	db.mu.Lock()
	if eng, ok := db.open[name]; ok {
		db.mu.Unlock()
		return option.Some(eng.engine), nil
	}
	closed := db.closed
	db.mu.Unlock()

	if closed {
		return option.None[Engine](), fmt.Errorf("ndb: database is closed")
	}

	dir := filepath.Join(db.root, name)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return option.None[Engine](), nil
		}
		return option.None[Engine](), fmt.Errorf("ndb: stat engine %q: %w", name, err)
	}

	eng, err := db.Engine(name, EngineOptions{})
	if err != nil {
		return option.None[Engine](), err
	}
	return option.Some(eng), nil
}

// Engines iterates over all engine instances present on disk, reading each
// instance's recorded kind. Instances are not opened.
func (db *DB) Engines() iter.Seq2[EngineInfo, error] {
	return func(yield func(EngineInfo, error) bool) {
		entries, err := os.ReadDir(db.root)
		if err != nil {
			yield(EngineInfo{}, fmt.Errorf("ndb: read root dir: %w", err))
			return
		}

		var names []string
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			names = append(names, e.Name())
		}
		slices.Sort(names)

		for _, name := range names {
			kind, err := readEngineMarker(filepath.Join(db.root, name))
			if err != nil {
				if !yield(EngineInfo{Name: name}, err) {
					return
				}
				continue
			}
			if !yield(EngineInfo{Name: name, Kind: kind}, nil) {
				return
			}
		}
	}
}

// Close closes all currently open engine instances, releasing their resources
// (file pools, locks, buffered writes). It is safe to call multiple times.
// After Close the DB handle is spent: further [DB.Engine] calls return an error.
//
// Each instance is released through the close function its factory returned, so
// that only the DB controls the instance lifecycle.
func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.closed = true

	var firstErr error
	for name, eng := range db.open {
		if eng.close != nil {
			if err := eng.close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		delete(db.open, name)
	}
	return firstErr
}

// resolveEngineKind determines which engine kind to use for the instance at dir:
// the recorded kind for an existing instance, otherwise the requested kind
// (want) and finally the DB's configured default for a new one. Caller must
// hold db.mu.
func (db *DB) resolveEngineKind(dir string, want EngineKind) (EngineKind, error) {
	recorded, err := readEngineMarker(dir)
	if err != nil {
		return "", err
	}
	if recorded != "" {
		return recorded, nil
	}
	if want != "" {
		return want, nil
	}
	return db.opts.DefaultKind, nil
}

func validateEngineName(name string) error {
	if name == "" {
		return fmt.Errorf("ndb: empty engine name")
	}
	if strings.ContainsAny(name, `/\`) || strings.HasPrefix(name, ".") {
		return fmt.Errorf("ndb: invalid engine name %q", name)
	}
	return nil
}

func readEngineMarker(dir string) (EngineKind, error) {
	data, err := os.ReadFile(filepath.Join(dir, engineMarker))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("ndb: read engine marker: %w", err)
	}
	return EngineKind(strings.TrimSpace(string(data))), nil
}

func writeEngineMarker(dir string, kind EngineKind) error {
	path := filepath.Join(dir, engineMarker)
	if existing, err := readEngineMarker(dir); err == nil && existing == kind {
		return nil
	}
	if err := os.WriteFile(path, []byte(kind), 0644); err != nil {
		return fmt.Errorf("ndb: write engine marker: %w", err)
	}
	return nil
}
