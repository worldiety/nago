// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"go.wdy.de/nago/pkg/ndb"
)

// NDBInstance describes one ndb database registered with the Configurator.
type NDBInstance struct {
	// Path is the resolved absolute filesystem path of the database.
	Path string
	// Name is a short, human-readable label derived from the path's base name
	// (e.g. "audit.ndb" for DataDir()/audit.ndb).
	Name string
	// DB is the open database handle.
	DB *ndb.DB
}

// NDB returns the shared application ndb database rooted at DataDir()/ndb.
// It is a convenience wrapper over [Configurator.OpenNDB]("ndb"). The instance
// is created on first use, cached, and closed automatically on application
// shutdown.
func (c *Configurator) NDB() (*ndb.DB, error) {
	return c.OpenNDB("ndb")
}

// OpenNDB opens (or returns the already-cached) ndb database at path. An
// absolute path is used verbatim; a relative path is resolved under DataDir(),
// mirroring the conventional "<name>.ndb" layout (e.g. OpenNDB("audit.ndb")
// resolves to DataDir()/audit.ndb).
//
// Databases are cached by their resolved absolute path, so repeated calls with
// an equivalent path return the same *ndb.DB. All databases opened through the
// Configurator are closed automatically on application shutdown.
//
// Each database owns its own file-descriptor pool: an *ndb.DB closes its pool on
// Close, so pools cannot be shared across databases.
func (c *Configurator) OpenNDB(path string) (*ndb.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("ndb: empty database path")
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(c.DataDir(), path)
	}

	c.ndbMutex.Lock()
	defer c.ndbMutex.Unlock()

	if db, ok := c.ndbs[path]; ok {
		return db, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), defaultDirPermission); err != nil {
		return nil, fmt.Errorf("ndb: cannot create parent dir for %q: %w", path, err)
	}

	db, err := ndb.Open(path, ndb.Options{})
	if err != nil {
		return nil, fmt.Errorf("ndb: cannot open %q: %w", path, err)
	}

	if c.ndbs == nil {
		c.ndbs = make(map[string]*ndb.DB)
		// Register a single destructor that closes every ndb database opened
		// through the Configurator when the application shuts down.
		c.OnDestroy(func() {
			c.ndbMutex.Lock()
			defer c.ndbMutex.Unlock()
			for p, db := range c.ndbs {
				if err := db.Close(); err != nil {
					slog.Error("cannot close ndb database", "path", p, "err", err.Error())
				}
			}
			c.ndbs = nil
		})
	}
	c.ndbs[path] = db

	return db, nil
}

// NDBInstances returns all ndb databases that have been opened through the
// Configurator (via [Configurator.NDB] or [Configurator.OpenNDB]), sorted by
// path. This lets tooling (e.g. the ndb inspector) discover and browse every
// registered database, not only the default one.
//
// Only databases that have actually been opened are returned; a database is
// opened lazily on its first NDB/OpenNDB call. Callers that need a specific
// database present should call OpenNDB for it first.
func (c *Configurator) NDBInstances() []NDBInstance {
	c.ndbMutex.Lock()
	defer c.ndbMutex.Unlock()

	out := make([]NDBInstance, 0, len(c.ndbs))
	for path, db := range c.ndbs {
		out = append(out, NDBInstance{
			Path: path,
			Name: filepath.Base(path),
			DB:   db,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}
