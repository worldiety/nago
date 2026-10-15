// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui_test

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/lib/ui/internal/cells"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	nui "go.wdy.de/nago/presentation/ui"
	nsheet "go.wdy.de/nago/presentation/ui/sheet"
)

// source creates the sheet of a page.
type source struct {
	name string
	view func(t testing.TB) func(wnd core.Window) core.View
}

var sources = []source{
	{"go", func(testing.TB) func(core.Window) core.View {
		header := nsheet.Style{BackgroundColor: "#eeeeee", Bold: true, Borders: nsheet.All(nsheet.Line{Color: "#999999", Style: nsheet.LineSolid})}
		return func(core.Window) core.View {
			return nsheet.Sheet("big", 50000, 30, func(row, col int) nsheet.Cell {
				if row == 0 || col == 0 {
					return nsheet.Cell{Text: strconv.Itoa(row + col), Style: header}
				}
				n := strconv.Itoa(row*30 + col)
				return nsheet.Cell{Text: n, Value: n, Type: nsheet.ValueNumber, Editor: nsheet.EditorNumber}
			})
		}
	}},
	{"native", func(testing.TB) func(core.Window) core.View {
		return func(core.Window) core.View { return cells.View().Core() }
	}},
	{"goja", func(t testing.TB) func(core.Window) core.View {
		in := load(t)
		return func(core.Window) core.View {
			res, err := in.Call(context.Background(), "View")
			if err != nil {
				t.Fatal(err)
			}
			var v ui.View
			if err := gostrict.FromNeutral(res[0], &v); err != nil {
				t.Fatal(err)
			}
			return v.Core()
		}
	}},
}

func load(t testing.TB) gostrict.Instance {
	b, diags, err := gostrict.Compile("internal/cells")
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags.Text())
	}
	in, err := new(gojaengine.Engine).Load(context.Background(), b, ui.Host())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = in.Close() })
	return in
}

func open(t testing.TB, src source) (*nagotest.Window, nagotest.Selection) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID(core.ApplicationID("de.worldiety.gostrict.ui." + src.name))
		page := src.view(t)
		cfg.RootView("page", func(wnd core.Window) core.View {
			return nui.VStack(page(wnd))
		})
	})
	w := app.Open(t, nil, "page")
	return w, w.Find(nagotest.Type[*proto.Sheet]())
}

// TestSheet renders the same cells in Go, natively through the header package and translated in goja, also after
// scrolling, when the window calls the function of the program without a call of the instance.
func TestSheet(t *testing.T) {
	var want []proto.SheetCells
	for _, src := range sources {
		w, s := open(t, src)
		w.SheetScroll(s, nsheet.Range{Row: 1000, Col: 3, Rows: 40, Cols: 10})
		got := []proto.SheetCells{w.Find(nagotest.Type[*proto.Sheet]()).Node().Component.(*proto.Sheet).Cells}
		if want == nil {
			want = got
			if len(want[0]) < 400 {
				t.Fatalf("only %d cells", len(want[0]))
			}
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s renders other cells than go", src.name)
		}
	}
}

// BenchmarkSheet measures a change of the window of 90 x 29 cells and the frozen header, which calls the function of
// the cells about 2700 times.
func BenchmarkSheet(b *testing.B) {
	for _, src := range sources {
		b.Run(src.name, func(b *testing.B) {
			w, s := open(b, src)
			i := 0
			start := time.Now()
			for b.Loop() {
				w.SheetScroll(s, nsheet.Range{Row: 1 + (i%90)*500, Col: 1, Rows: 90, Cols: 29})
				i++
			}
			cells := len(w.Find(nagotest.Type[*proto.Sheet]()).Node().Component.(*proto.Sheet).Cells)
			b.ReportMetric(float64(time.Since(start).Nanoseconds())/float64(i)/float64(cells)*1000, "ns/1000cells")
		})
	}
}

// BenchmarkCallback measures the calls of the function of the cells from Go into goja and the conversion of the
// cells, without nago.
func BenchmarkCallback(b *testing.B) {
	var cell gostrict.Callback
	host := ui.Host()
	sheetFn := host["go.wdy.de/nago/pkg/gostrict/lib/ui.Sheet"]
	host["go.wdy.de/nago/pkg/gostrict/lib/ui.Sheet"] = func(args []any) []any {
		cell = args[3].(gostrict.Callback)
		return sheetFn(args)
	}
	bnd, _, err := gostrict.Compile("internal/cells")
	if err != nil {
		b.Fatal(err)
	}
	in, err := new(gojaengine.Engine).Load(context.Background(), bnd, host)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := in.Call(context.Background(), "View"); err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		for row := range int32(40) {
			for col := range int32(25) {
				if _, err := cell.Call(row, col); err != nil {
					b.Fatal(err)
				}
			}
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N), "ns/1000cells")
}
