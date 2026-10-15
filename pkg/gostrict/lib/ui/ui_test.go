// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui_test

import (
	"context"
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/lib/ui/internal/agenda"
	"go.wdy.de/nago/pkg/gostrict/lib/ui/internal/cells"
	"go.wdy.de/nago/pkg/gostrict/lib/ui/internal/extras"
	"go.wdy.de/nago/pkg/gostrict/lib/ui/internal/form"
	"go.wdy.de/nago/pkg/gostrict/lib/ui/internal/page"
	"go.wdy.de/nago/pkg/gostrict/lib/ui/internal/people"
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
		return func(core.Window) core.View { return ui.Core(cells.View()) }
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
			return ui.Core(v)
		}
	}},
}

func load(t testing.TB) gostrict.Instance {
	return loadDir(t, "internal/cells")
}

func loadDir(t testing.TB, dir string) gostrict.Instance {
	b, diags, err := gostrict.Compile(dir)
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

// TestVars compares the bindings of the variables of the package with their values in Go.
func TestVars(t *testing.T) {
	prog, err := gostrict.Load(".")
	if err != nil {
		t.Fatal(err)
	}
	vars, records := gostrict.HeaderVars(prog)
	if len(vars) == 0 {
		t.Fatal("no variables")
	}
	for _, m := range gojaengine.CheckVars(vars, records, ui.HostVars()) {
		t.Error(m)
	}
}

// TestPage renders a window of a program natively and in goja and clicks a button, which changes a state of the
// window, so that both render the same views.
func TestPage(t *testing.T) {
	in := loadDir(t, "internal/page")
	renders := map[string]func(wnd core.Window) core.View{
		"native": func(wnd core.Window) core.View { return ui.Core(page.Render(ui.WindowOf(wnd))) },
		"goja":   func(wnd core.Window) core.View { return ui.Render(context.Background(), in, "Render", wnd) },
	}
	var want []string
	for _, name := range []string{"native", "goja"} {
		app := nagotest.New(t, func(cfg *application.Configurator) {
			cfg.SetApplicationID(core.ApplicationID("de.worldiety.gostrict.page." + name))
			cfg.RootView("page", renders[name])
		})
		w := app.Open(t, nil, "page")
		var got []string
		got = append(got, treeOf(t, w))
		w.Find(nagotest.Text("zu"))
		w.Click(w.Find(nagotest.Text("Umschalten")))
		w.Find(nagotest.Text("offen"))
		got = append(got, treeOf(t, w))
		if want == nil {
			want = got
			continue
		}
		for i := range want {
			if normalize(got[i]) != normalize(want[i]) {
				t.Fatalf("%s renders another tree %d:\n%s", name, i, firstDiff(got[i], want[i]))
			}
		}
	}
}

// treeOf returns the whole rendered tree of the window as JSON, with all nested components.
func treeOf(t testing.TB, w *nagotest.Window) string {
	t.Helper()
	buf, err := json.Marshal(w.Tree())
	if err != nil {
		t.Fatal(err)
	}
	return string(buf)
}

// pointers are the numbers of callbacks and states and generated ids, which are random and differ between two
// windows.
var (
	pointers = regexp.MustCompile(`:[0-9]{7,}`)
	ids      = regexp.MustCompile(`"([a-z]+)[0-9]{7,}"`)
)

func normalize(s string) string {
	return ids.ReplaceAllString(pointers.ReplaceAllString(s, ":PTR"), `"${1}ID"`)
}

// firstDiff returns the context of the first difference of a and b.
func firstDiff(a, b string) string {
	a, b = normalize(a), normalize(b)
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	from := max(i-120, 0)
	return "got: ..." + a[from:min(i+80, len(a))] + "\nwant: ..." + b[from:min(i+80, len(b))]
}

// TestForm renders a form with many views natively and in goja, types into its fields and clicks a button, so that
// both render the same views after each step.
func TestForm(t *testing.T) {
	in := loadDir(t, "internal/form")
	renders := map[string]func(wnd core.Window) core.View{
		"native": func(wnd core.Window) core.View { return ui.Core(form.Render(ui.WindowOf(wnd))) },
		"goja":   func(wnd core.Window) core.View { return ui.Render(context.Background(), in, "Render", wnd) },
	}
	var want []string
	for _, name := range []string{"native", "goja"} {
		app := nagotest.New(t, func(cfg *application.Configurator) {
			cfg.SetApplicationID(core.ApplicationID("de.worldiety.gostrict.form." + name))
			cfg.RootView("form", renders[name])
		})
		w := app.Open(t, nil, "form")
		got := []string{treeOf(t, w)}
		w.Find(nagotest.Text("Hallo  (0)"))

		w.Type(w.Find(nagotest.Label("Name")), "Anna")
		w.Find(nagotest.Text("Hallo Anna (0)"))
		got = append(got, treeOf(t, w))

		w.Click(w.Find(nagotest.Text("Älter")))
		w.Type(w.Find(nagotest.Type[*proto.Checkbox]()), "true")
		w.Find(nagotest.Text("Hallo Anna (1), einverstanden"))
		got = append(got, treeOf(t, w))

		w.Type(w.Find(nagotest.Label("Ort")), "ol")
		w.Find(nagotest.Text("Hallo Anna (1), einverstanden aus ol"))
		got = append(got, treeOf(t, w))

		if want == nil {
			want = got
			continue
		}
		for i := range want {
			if normalize(got[i]) != normalize(want[i]) {
				t.Fatalf("%s renders another tree %d:\n%s", name, i, firstDiff(got[i], want[i]))
			}
		}
	}
}

// TestDataView shows the rows of a program in a data view natively and in goja, which render the same views.
func TestDataView(t *testing.T) {
	in := loadDir(t, "internal/people")
	renders := map[string]func(wnd core.Window) core.View{
		"native": func(wnd core.Window) core.View { return ui.Core(people.Render(ui.WindowOf(wnd))) },
		"goja":   func(wnd core.Window) core.View { return ui.Render(context.Background(), in, "Render", wnd) },
	}
	var want string
	for _, name := range []string{"native", "goja"} {
		app := nagotest.New(t, func(cfg *application.Configurator) {
			cfg.SetApplicationID(core.ApplicationID("de.worldiety.gostrict.people." + name))
			cfg.RootView("people", renders[name])
		})
		w := app.Open(t, nil, "people")
		w.Find(nagotest.Text("Anna"))
		w.Find(nagotest.Text("Carla"))
		got := normalize(treeOf(t, w))
		if want == "" {
			want = got
			continue
		}
		if got != want {
			t.Fatalf("%s renders another data view:\n%s", name, firstDiff(got, want))
		}
	}
}

// TestAgenda renders a calendar, pickers for dates and times and a flowchart natively and in goja, which render the
// same views.
func TestAgenda(t *testing.T) {
	in := loadDir(t, "internal/agenda")
	renders := map[string]func(wnd core.Window) core.View{
		"native": func(wnd core.Window) core.View { return ui.Core(agenda.Render(ui.WindowOf(wnd))) },
		"goja":   func(wnd core.Window) core.View { return ui.Render(context.Background(), in, "Render", wnd) },
	}
	var want string
	for _, name := range []string{"native", "goja"} {
		app := nagotest.New(t, func(cfg *application.Configurator) {
			cfg.SetApplicationID(core.ApplicationID("de.worldiety.gostrict.agenda." + name))
			cfg.RootView("agenda", renders[name])
		})
		w := app.Open(t, nil, "agenda")
		w.Find(nagotest.Text("Kalender"))
		w.Find(nagotest.Text("Bescherung"))
		w.Find(nagotest.Text("Spaziergang"))
		got := normalize(treeOf(t, w))
		if want == "" {
			want = got
			continue
		}
		if got != want {
			t.Fatalf("%s renders another tree:\n%s", name, firstDiff(got, want))
		}
	}
}

// TestExtras renders the picker, a background, embedded images, icons and the colors of a chart natively and in
// goja, picks rows by a button, and compares the views.
func TestExtras(t *testing.T) {
	in := loadDir(t, "internal/extras")
	renders := map[string]func(wnd core.Window) core.View{
		"native": func(wnd core.Window) core.View { return ui.Core(extras.Render(ui.WindowOf(wnd))) },
		"goja":   func(wnd core.Window) core.View { return ui.Render(context.Background(), in, "Render", wnd) },
	}
	var want []string
	for _, name := range []string{"native", "goja"} {
		app := nagotest.New(t, func(cfg *application.Configurator) {
			cfg.SetApplicationID(core.ApplicationID("de.worldiety.gostrict.extras." + name))
			cfg.RootView("extras", renders[name])
		})
		w := app.Open(t, nil, "extras")
		got := []string{normalize(treeOf(t, w))}
		for _, part := range []string{"#ff0000", "linear-gradient", "PHN2Zy"} {
			if !strings.Contains(got[0], part) {
				t.Errorf("%s lacks %s", name, part)
			}
		}
		if false {
			t.Fatalf("%s lacks the colors, the background or an icon", name)
		}
		w.Click(w.Find(nagotest.Text("Speichern")))
		w.Find(nagotest.Text("gewählt: Apfel, Kirsche"))
		got = append(got, normalize(treeOf(t, w)))
		if want == nil {
			want = got
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s renders another tree %d:\n%s", name, i, firstDiff(got[i], want[i]))
			}
		}
	}
}
