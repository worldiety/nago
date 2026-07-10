// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/tsdb" // registers the "tsdb" engine
)

func TestNDBIntegration(t *testing.T) {
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{}))
	defer func() { _ = db.Close() }()

	eng, err := db.Engine("metrics", ndb.EngineOptions{Kind: "tsdb", Config: tsdb.Options{}})
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	if eng.Kind() != "tsdb" {
		t.Fatalf("kind = %q", eng.Kind())
	}

	se, ok := eng.(ndb.SeriesEngine)
	if !ok {
		t.Fatal("engine does not implement SeriesEngine")
	}

	tse := eng.(interface{ DB() *tsdb.DB })
	c, err := tse.DB().Column("host1", "cpu", tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 2})
	if err != nil {
		t.Fatal(err)
	}
	for i := int64(0); i < 500; i++ {
		if err := c.PutF64(1000+i*100, 0.5+float64(i)/1000); err != nil {
			t.Fatal(err)
		}
	}
	if err := c.Flush(); err != nil {
		t.Fatal(err)
	}

	cols, err := se.SeriesColumns()
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 1 || cols[0] != "host1/cpu" {
		t.Fatalf("SeriesColumns = %v", cols)
	}

	var n int
	err = c.ScanF64(0, 1<<62, func(ts []int64, vals []float64) bool {
		n += len(ts)
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 500 {
		t.Fatalf("read %d points, want 500", n)
	}
}
