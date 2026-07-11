// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndbinspector_test

import (
	"testing"

	"github.com/worldiety/option"
	ndbinspector "go.wdy.de/nago/application/inspector/ndb"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/tsdb"
)

func TestUseCasesTimeseries(t *testing.T) {
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{}))
	defer db.Close()

	// seed a numeric and a string column via the tsdb engine
	eng := option.Must(db.Engine("metrics", ndb.EngineOptions{Kind: tsdb.EngineKind, Config: tsdb.Options{}}))
	tdb := eng.(interface{ DB() *tsdb.DB }).DB()

	base := int64(1_700_000_000_000)
	num := option.Must(tdb.Column("plant", "temp", tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 2}))
	for i := int64(0); i < 5000; i++ {
		if err := num.PutF64(base+i*20, 20.0+float64(i%7)); err != nil {
			t.Fatal(err)
		}
	}
	if err := num.Flush(); err != nil {
		t.Fatal(err)
	}

	str := option.Must(tdb.Column("plant", "state", tsdb.Schema{Scheme: tsdb.SchemeString}))
	for i := int64(0); i < 300; i++ {
		if err := str.PutString(base+i*1000, "running"); err != nil {
			t.Fatal(err)
		}
	}
	if err := str.Flush(); err != nil {
		t.Fatal(err)
	}

	const instPath = "/db/main"
	uc := ndbinspector.NewUseCases(func() []ndbinspector.Instance {
		return []ndbinspector.Instance{{Path: instPath, Name: "main", DB: db}}
	})
	su := user.SU()

	// engines
	engines, err := uc.SeriesEngines(su, instPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(engines) != 1 || engines[0].Name != "metrics" || engines[0].Kind != tsdb.EngineKind {
		t.Fatalf("engines = %+v", engines)
	}

	// columns + stats
	cols, err := uc.Columns(su, instPath, "metrics")
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 2 {
		t.Fatalf("want 2 columns, got %+v", cols)
	}
	var temp, state ndbinspector.ColumnInfo
	for _, ci := range cols {
		switch ci.Column {
		case "temp":
			temp = ci
		case "state":
			state = ci
		}
	}
	if !temp.Numeric() || temp.Scheme != tsdb.SchemeDecimal || temp.Decimals != 2 {
		t.Fatalf("temp column wrong: %+v", temp)
	}
	if !temp.HasData || temp.MinMillis != base || temp.MaxMillis != base+4999*20 {
		t.Fatalf("temp range wrong: %+v", temp)
	}
	if temp.Chunks < 1 || temp.Bytes <= 0 {
		t.Fatalf("temp chunks/bytes wrong: %+v", temp)
	}
	if n, err := uc.CountColumn(su, instPath, "metrics", "plant", "temp"); err != nil || n != 5000 {
		t.Fatalf("temp count = %d (err %v), want 5000", n, err)
	}
	if state.Numeric() || state.Scheme != tsdb.SchemeString {
		t.Fatalf("state column wrong: %+v", state)
	}
	if n, err := uc.CountColumn(su, instPath, "metrics", "plant", "state"); err != nil || n != 300 {
		t.Fatalf("state count = %d (err %v), want 300", n, err)
	}

	// M4 over the numeric column is downsampled and bounded
	pts, err := uc.SeriesM4(su, ndbinspector.SeriesRequest{
		Instance: instPath, Engine: "metrics", Bucket: "plant", Column: "temp",
		MinMillis: temp.MinMillis, MaxMillis: temp.MaxMillis, Width: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) == 0 || len(pts) > 4*100 {
		t.Fatalf("M4 point count out of bounds: %d", len(pts))
	}
	for i := 1; i < len(pts); i++ {
		if pts[i].Millis <= pts[i-1].Millis {
			t.Fatalf("M4 points not ascending at %d", i)
		}
	}

	// M4 on a string column is rejected
	if _, err := uc.SeriesM4(su, ndbinspector.SeriesRequest{
		Instance: instPath, Engine: "metrics", Bucket: "plant", Column: "state",
		MinMillis: base, MaxMillis: base + 1000, Width: 50,
	}); err == nil {
		t.Fatal("M4 on a string column must fail")
	}

	// string window respects the limit
	rows, err := uc.StringWindow(su, ndbinspector.StringWindowRequest{
		Instance: instPath, Engine: "metrics", Bucket: "plant", Column: "state",
		MinMillis: base, MaxMillis: base + 300*1000, Limit: 50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 50 {
		t.Fatalf("string window limit not enforced: %d", len(rows))
	}
	if rows[0].Value != "running" {
		t.Fatalf("string value = %q", rows[0].Value)
	}

	// delete a range in the numeric column and confirm it shrinks
	if err := uc.DeleteSeriesRange(su, instPath, "metrics", "plant", "temp", base, base+1000*20); err != nil {
		t.Fatal(err)
	}
	if err := uc.FlushColumn(su, instPath, "metrics", "plant", "temp"); err != nil {
		t.Fatal(err)
	}
	after, err := uc.SeriesM4(su, ndbinspector.SeriesRequest{
		Instance: instPath, Engine: "metrics", Bucket: "plant", Column: "temp",
		MinMillis: base, MaxMillis: base + 1000*20, Width: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 0 {
		t.Fatalf("deleted range should be empty, got %d points", len(after))
	}
}
