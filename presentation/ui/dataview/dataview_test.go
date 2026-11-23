// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataview_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/pager"
)

type personID string

type person struct {
	ID   personID
	Name string
}

func (p person) Identity() personID {
	return p.ID
}

// people contains persons without a name, which must stay last in both directions and in their original order.
var people = []person{{"a", "Bert"}, {"b", ""}, {"c", "Anna"}, {"d", "Carl"}, {"e", ""}}

func peopleView(wnd core.Window) dataview.TDataView[person, personID] {
	var ids []personID
	for _, p := range people {
		ids = append(ids, p.ID)
	}

	return dataview.FromData(wnd, dataview.Data[person, personID]{
		ID: "people",
		FindAll: func(yield func(personID, error) bool) {
			for _, id := range ids {
				if !yield(id, nil) {
					return
				}
			}
		},
		FindByID: func(id personID) (option.Opt[person], error) {
			for _, p := range people {
				if p.ID == id {
					return option.Some(p), nil
				}
			}

			return option.None[person](), nil
		},
		Fields: []dataview.Field[person]{
			{
				Name:       "Name",
				Map:        func(p person) core.View { return ui.Text("row " + string(p.ID)) },
				Comparator: func(a, b person) int { return strings.Compare(a.Name, b.Name) },
				Empty:      func(p person) bool { return p.Name == "" },
			},
			{
				// a field without a name gets no sort control
				Map:        func(p person) core.View { return ui.Text(p.Name) },
				Comparator: func(a, b person) int { return strings.Compare(a.Name, b.Name) },
			},
		},
	}).Search(true).CreateAction(func() {})
}

func configure(cfg *application.Configurator) {
	cfg.SetApplicationID("de.worldiety.dataview")

	cfg.RootView("table", func(wnd core.Window) core.View {
		return peopleView(wnd).Style(dataview.Table)
	})

	cfg.RootView("cards", func(wnd core.Window) core.View {
		return peopleView(wnd).Style(dataview.Card)
	})

	cfg.RootView("paged", func(wnd core.Window) core.View {
		return peopleView(wnd).Style(dataview.Table).ModelOptions(pager.ModelOptions{PageSize: 3})
	})

	for _, style := range []struct {
		name  string
		style dataview.Style
	}{{"table", dataview.Table}, {"cards", dataview.Card}, {"list", dataview.List}} {
		cfg.RootView(core.NavigationPath("paged-"+style.name), func(wnd core.Window) core.View {
			return peopleView(wnd).Style(style.style).ModelOptions(pager.ModelOptions{PageSize: 3})
		})
	}

	cfg.RootView("wide", func(wnd core.Window) core.View {
		return peopleView(wnd).Style(dataview.Table).TableOptions(dataview.TableOptions{MinWidth: "1800px"})
	})
}

func rows(w *nagotest.Window) []string {
	var res []string
	for _, n := range w.FindAll(nagotest.TextContains("row ")).Nodes() {
		res = append(res, strings.TrimPrefix(string(n.Component.(*proto.TextView).Value), "row "))
	}

	return res
}

func sortControls(w *nagotest.Window) nagotest.Selection {
	return w.FindAll(nagotest.Where("sort control", func(n nagotest.Node) bool {
		c, ok := n.Component.(*proto.Stack)
		return ok && strings.Contains(string(c.AccessibilityLabel), "Sort")
	}))
}

func TestSortKeepsEmptyLast(t *testing.T) {
	for _, route := range []core.NavigationPath{"table", "cards"} {
		t.Run(string(route), func(t *testing.T) {
			app := nagotest.New(t, configure)
			w := app.Open(t, nil, route, nagotest.Locale("en"))

			sortControls(w).Exactly(1)

			w.Click(w.Find(nagotest.Label("Sort by Name")))
			if got, want := rows(w), []string{"c", "a", "d", "b", "e"}; !slices.Equal(got, want) {
				t.Fatalf("ascending: got %v, want %v", got, want)
			}

			w.Click(w.Find(nagotest.Label("Sorted by Name ascending")))
			if got, want := rows(w), []string{"d", "a", "c", "b", "e"}; !slices.Equal(got, want) {
				t.Fatalf("descending: got %v, want %v", got, want)
			}

			w.Find(nagotest.Label("Sorted by Name descending"))
		})
	}
}

func TestPagerIsLabelled(t *testing.T) {
	app := nagotest.New(t, configure)
	w := app.Open(t, nil, "paged", nagotest.Locale("en"))

	w.Find(nagotest.Text("1 of 2"))
	w.Click(w.Find(nagotest.Label("Next page, currently page 1 of 2")))
	w.Find(nagotest.Text("2 of 2"))
	w.Find(nagotest.Label("Previous page, currently page 2 of 2")).Enabled()
}

func TestMinWidthScrollsOnlyTheTable(t *testing.T) {
	app := nagotest.New(t, configure)
	w := app.Open(t, nil, "wide", nagotest.Locale("en"))

	scroll := w.Find(nagotest.Type[*proto.ScrollView]())
	if sv := scroll.Node().Component.(*proto.ScrollView); sv.Axis != proto.ScrollViewAxisHorizontal {
		t.Fatalf("unexpected axis %v", sv.Axis)
	}

	table := scroll.Find(nagotest.Type[*proto.Table]())
	if got := table.Node().Component.(*proto.Table).Frame.MinWidth; got != "1800px" {
		t.Fatalf("unexpected min width %q", got)
	}

	// the action bar and the pager stay outside the scrolled area
	w.Find(nagotest.Type[*proto.TextField]())
	scroll.FindAll(nagotest.Type[*proto.TextField]()).None()
	scroll.FindAll(nagotest.Label("New")).None()
	w.Find(nagotest.Text("5 entries"))
	scroll.FindAll(nagotest.Text("5 entries")).None()
}

func TestPagerKeepsFocusAndTellsPage(t *testing.T) {
	for _, style := range []string{"table", "cards", "list"} {
		t.Run(style, func(t *testing.T) {
			app := nagotest.New(t, configure)
			w := app.Open(t, nil, core.NavigationPath("paged-"+style), nagotest.Locale("en"))

			next := w.Find(nagotest.Label("Next page, currently page 1 of 2"))
			nextID := string(next.Node().Component.(*proto.Stack).Id)
			if nextID == "" {
				t.Fatal("the next button needs a stable id to keep the focus")
			}

			w.Click(next)

			// the label of the focused button tells the new page
			next = w.Find(nagotest.Label("Next page, currently page 2 of 2"))
			if got := string(next.Node().Component.(*proto.Stack).Id); got != nextID {
				t.Fatalf("the id of the next button changed from %q to %q", nextID, got)
			}

			// the next button is disabled on the last page, thus the focus moves to the previous button
			prev := w.Find(nagotest.Label("Previous page, currently page 2 of 2"))
			if got, want := w.Focused(), string(prev.Node().Component.(*proto.Stack).Id); got != want {
				t.Fatalf("focused %q, want %q", got, want)
			}
		})
	}
}

func TestEntriesAreNamedRegion(t *testing.T) {
	for _, style := range []string{"table", "cards", "list"} {
		t.Run(style, func(t *testing.T) {
			app := nagotest.New(t, configure)
			w := app.Open(t, nil, core.NavigationPath("paged-"+style), nagotest.Locale("en"))

			region := w.Find(nagotest.Where("entries region", func(n nagotest.Node) bool {
				c, ok := n.Component.(*proto.Accessible)
				return ok && c.Role == "region" && c.Label == "Entries"
			}))

			region.Find(nagotest.Text("row a"))
			// the region starts with the entries, the search belongs to the action bar in front of it
			region.FindAll(nagotest.Type[*proto.TextField]()).None()
		})
	}
}
