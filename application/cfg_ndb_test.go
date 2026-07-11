// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"path/filepath"
	"testing"

	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

func TestConfiguratorOpenNDB(t *testing.T) {
	root := t.TempDir()
	c := &Configurator{dataDir: root}

	// relative path resolves under DataDir()
	dbA, err := c.OpenNDB("audit.ndb")
	if err != nil {
		t.Fatalf("open a: %v", err)
	}

	// second call with the equivalent path returns the same cached instance
	dbA2, err := c.OpenNDB("audit.ndb")
	if err != nil {
		t.Fatalf("open a again: %v", err)
	}
	if dbA != dbA2 {
		t.Fatal("expected the same cached *ndb.DB for the same path")
	}

	// absolute path is used verbatim and is a distinct instance
	absPath := filepath.Join(root, "other.ndb")
	dbB, err := c.OpenNDB(absPath)
	if err != nil {
		t.Fatalf("open b: %v", err)
	}
	if dbB == dbA {
		t.Fatal("distinct paths must yield distinct databases")
	}

	// the returned DB is usable: create an engine on it
	eng, err := dbA.Engine("events", ndb.EngineOptions{Kind: msgstore.EngineKind, Config: msgstore.Options{}})
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	if eng.Kind() != msgstore.EngineKind {
		t.Fatalf("kind = %q", eng.Kind())
	}

	// NDB() is the default, rooted at DataDir()/ndb, distinct from the above.
	dbDefault, err := c.NDB()
	if err != nil {
		t.Fatalf("default ndb: %v", err)
	}
	if dbDefault == dbA || dbDefault == dbB {
		t.Fatal("default NDB must be its own instance")
	}

	// NDBInstances lists every opened database, sorted by path.
	insts := c.NDBInstances()
	if len(insts) != 3 {
		t.Fatalf("want 3 registered instances, got %d: %+v", len(insts), insts)
	}
	byPath := map[string]NDBInstance{}
	for _, in := range insts {
		byPath[in.Path] = in
	}
	if in, ok := byPath[absPath]; !ok || in.Name != "other.ndb" || in.DB != dbB {
		t.Fatalf("other.ndb instance wrong: %+v", byPath[absPath])
	}
	if in, ok := byPath[filepath.Join(root, "audit.ndb")]; !ok || in.DB != dbA {
		t.Fatalf("audit.ndb instance wrong: %+v", byPath[filepath.Join(root, "audit.ndb")])
	}

	// closing all cached databases (as the shutdown destructor does) must succeed.
	for _, db := range c.ndbs {
		if err := db.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
	}
}

func TestConfiguratorOpenNDBEmptyPath(t *testing.T) {
	c := &Configurator{dataDir: t.TempDir()}
	if _, err := c.OpenNDB(""); err == nil {
		t.Fatal("empty path must be rejected")
	}
}
