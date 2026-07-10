// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"fmt"

	"go.wdy.de/nago/pkg/ndb"
)

// EngineKind is the [ndb.EngineKind] under which the tsdb engine registers.
const EngineKind ndb.EngineKind = "tsdb"

func init() {
	ndb.Register(EngineKind, openEngine)
}

// engine adapts a *DB to the ndb.Engine / ndb.SeriesEngine contracts.
type engine struct {
	name string
	db   *DB
}

var (
	_ ndb.Engine       = (*engine)(nil)
	_ ndb.SeriesEngine = (*engine)(nil)
)

// openEngine is the ndb.EngineFactory for tsdb. cfg accepts nil (defaults) or an
// Options value; the shared FilePool is injected by ndb.
func openEngine(name, dir string, pool *ndb.FilePool, cfg ndb.EngineConfig) (ndb.Engine, func() error, error) {
	var opts Options
	switch c := cfg.(type) {
	case nil:
	case Options:
		opts = c
	default:
		return nil, nil, fmt.Errorf("tsdb: unsupported engine config type %T", cfg)
	}
	if opts.FilePool == nil {
		opts.FilePool = pool
	}
	db, err := Open(dir, opts)
	if err != nil {
		return nil, nil, err
	}
	return &engine{name: name, db: db}, db.Close, nil
}

func (e *engine) Name() string                     { return e.name }
func (e *engine) Kind() ndb.EngineKind             { return EngineKind }
func (e *engine) SeriesColumns() ([]string, error) { return e.db.SeriesColumns() }

// DB exposes the underlying tsdb handle for the full typed API. Do not Close it
// yourself: its lifecycle is owned by the ndb.DB that opened this instance.
func (e *engine) DB() *DB { return e.db }
