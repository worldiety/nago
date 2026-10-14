// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sheet_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui/sheet"
)

// The application sorts and filters its rows in its own model, the sheet only shows the result and the header.
func TestSheetSortAndFilter(t *testing.T) {
	names := []string{"Birne", "Apfel", "Kirsche", "Banane"}
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.sheettest")
		cfg.RootView("sheet", func(wnd core.Window) core.View {
			sorting := core.StateOf[sheet.Sorting](wnd, "sorting")
			filter := core.StateOf[string](wnd, "filter")

			rows := slices.Clone(names)
			rows = slices.DeleteFunc(rows, func(s string) bool { return !strings.Contains(strings.ToLower(s), filter.Get()) })
			if sorting.Get().Dir != sheet.SortNone {
				slices.Sort(rows)
				if sorting.Get().Dir == sheet.SortDesc {
					slices.Reverse(rows)
				}
			}

			return sheet.Sheet("fruits", len(rows)+1, 1, func(row, col int) sheet.Cell {
				if row == 0 {
					return sheet.Cell{Text: "Name"}
				}
				return sheet.Cell{Text: rows[row-1]}
			}).
				Frozen(1, 0).
				Sorted(sorting.Get()).
				OnSort(func(s sheet.Sorting) { sorting.Set(s) }).
				Filtered(map[int]string{0: filter.Get()}).
				OnFilter(func(col int, text string) { filter.Set(strings.ToLower(text)) })
		})
	})

	w := app.Open(t, nil, "sheet")
	s := w.Find(nagotest.Type[*proto.Sheet]())
	column := func() []string {
		sh := sheetOf(t, w)
		var res []string
		for row := 1; row < int(sh.RowCount); row++ {
			text, _, _ := cellText(sh, row, 0)
			res = append(res, text)
		}
		return res
	}

	w.SheetSort(s, sheet.Sorting{Col: 0, Dir: sheet.SortDesc})
	if got := column(); !slices.Equal(got, []string{"Kirsche", "Birne", "Banane", "Apfel"}) {
		t.Fatalf("unexpected sorting %v", got)
	}

	if h := sheetOf(t, w).Header; h.SortDir != proto.SheetSortDesc || h.SortCol != 0 {
		t.Fatalf("expected the header to show the sorting, got %+v", h)
	}

	w.SheetFilter(s, 0, "B")
	if got := column(); !slices.Equal(got, []string{"Birne", "Banane"}) {
		t.Fatalf("unexpected filter %v", got)
	}

	if f := sheetOf(t, w).Header.Filters; len(f) != 1 || f[0].Text != "b" {
		t.Fatalf("expected the header to show the filter, got %+v", f)
	}

	w.SheetFilter(s, 0, "")
	if got := column(); len(got) != 4 {
		t.Fatalf("expected the filter to be removed, got %v", got)
	}
}

func TestSheetSortRestricted(t *testing.T) {
	var sorted []sheet.Sorting
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.sheettest")
		cfg.RootView("sheet", func(wnd core.Window) core.View {
			return sheet.Sheet("s", 3, 3, func(row, col int) sheet.Cell { return sheet.Cell{Text: strconv.Itoa(col)} }).
				Frozen(1, 0).
				SortableColumns(1).
				OnSort(func(s sheet.Sorting) { sorted = append(sorted, s) })
		})
	})

	w := app.Open(t, nil, "sheet")
	s := w.Find(nagotest.Type[*proto.Sheet]())
	w.SheetSort(s, sheet.Sorting{Col: 0, Dir: sheet.SortAsc})
	w.SheetSort(s, sheet.Sorting{Col: 1, Dir: sheet.SortAsc})
	if len(sorted) != 1 || sorted[0].Col != 1 {
		t.Fatalf("expected only the sortable column, got %+v", sorted)
	}
}

// The user may show a single row or column again, which the user has hidden, but not one of the application.
func TestSheetUserHidden(t *testing.T) {
	m := &values{v: make([]int, 10)}
	w := openConfigured(t, m, func(s sheet.TSheet, _ core.Window) sheet.TSheet { return s.HiddenRows(9) })
	s := w.Find(nagotest.Type[*proto.Sheet]())

	w.SheetLayout(s, `{"op":"hideRow","index":2}`)
	w.SheetLayout(s, `{"op":"hideRow","index":3}`)
	if got := ints(sheetOf(t, w).UserHidden.Rows); !slices.Equal(got, []int{2, 3}) {
		t.Fatalf("expected the rows hidden by the user, got %v", got)
	}

	w.SheetLayout(s, `{"op":"showRow","index":2}`)
	sh := sheetOf(t, w)
	if !slices.Equal(ints(sh.HiddenRows), []int{3, 9}) || !slices.Equal(ints(sh.UserHidden.Rows), []int{3}) {
		t.Fatalf("unexpected hidden %v and user hidden %v", sh.HiddenRows, sh.UserHidden.Rows)
	}
}
