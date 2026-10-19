// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm_test

import (
	"context"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
)

const sheetFn = "go.wdy.de/nago/pkg/gostrict/lib/ui.Sheet"

// loadCells loads the program of the cells and returns the callback of the cells, which the host receives.
func loadCells(t testing.TB, e gostrict.Engine) (gostrict.Instance, *gostrict.Callback) {
	var cell gostrict.Callback
	host := ui.Host()
	sheet := host[sheetFn]
	host[sheetFn] = func(args []any) []any {
		cell = args[3].(gostrict.Callback)
		return sheet(args)
	}
	in, err := e.Load(context.Background(), compile(t, "../../lib/ui/internal/cells"), host)
	if err != nil {
		t.Fatal(err)
	}
	res, err := in.Call(context.Background(), "View")
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := res[0].(gostrict.Variant); !ok || v.Tag != "go.wdy.de/nago/pkg/gostrict/lib/ui.TSheet" {
		t.Fatalf("View returned %#v", res[0])
	}
	return in, &cell
}

func TestCells(t *testing.T) {
	_, cell := loadCells(t, &vm.Engine{})
	_, ref := loadCells(t, &gojaengine.Engine{})
	for _, rc := range [][2]int32{{0, 0}, {0, 5}, {3, 0}, {7, 9}, {1000, 29}} {
		got, err := (*cell).Call(rc[0], rc[1])
		if err != nil {
			t.Fatal(err)
		}
		want, err := (*ref).Call(rc[0], rc[1])
		if err != nil {
			t.Fatal(err)
		}
		if !gostrict.SameValues(got, want) {
			t.Fatalf("cell %v\n  vm:     %#v\n  native: %#v", rc, got, want)
		}
	}
}

func benchCallback(b *testing.B, e gostrict.Engine) {
	_, cell := loadCells(b, e)
	b.ReportAllocs()
	for b.Loop() {
		for row := range int32(40) {
			for col := range int32(25) {
				if _, err := (*cell).Call(row, col); err != nil {
					b.Fatal(err)
				}
			}
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/1000, "ns/cell")
}

// BenchmarkCallback calls the function of the cells 1000 times from Go and converts the cells, like the benchmark
// of the same name in lib/ui.
func BenchmarkCallback(b *testing.B) {
	b.Run("goja", func(b *testing.B) { benchCallback(b, &gojaengine.Engine{}) })
	b.Run("vm", func(b *testing.B) { benchCallback(b, &vm.Engine{}) })
}
