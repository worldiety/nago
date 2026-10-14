// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sheet_test

import (
	"slices"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/sheet"
)

func openConfigured(t *testing.T, m *values, configure func(s sheet.TSheet, wnd core.Window) sheet.TSheet) *nagotest.Window {
	t.Helper()
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.sheettest")
		cfg.RootView("sheet", func(wnd core.Window) core.View {
			return configure(sheet.Sheet("values", len(m.v), 2, m.cell).OnEdit(m.edit), wnd)
		})
	})

	return app.Open(t, nil, "sheet")
}

func ints(v proto.Ints) []int {
	var res []int
	for _, i := range v {
		res = append(res, int(i))
	}
	return res
}

// The user resizes and hides columns and rows and shows them again, but not the ones hidden by the application.
// The layout is kept in the bound state, e.g. to store it in the settings of the user.
func TestSheetLayout(t *testing.T) {
	m := &values{v: make([]int, 10)}
	var layout *core.State[sheet.Layout]
	w := openConfigured(t, m, func(s sheet.TSheet, wnd core.Window) sheet.TSheet {
		layout = core.StateOf[sheet.Layout](wnd, "layout")
		return s.HiddenRows(9).Layout(layout)
	})
	s := w.Find(nagotest.Type[*proto.Sheet]())

	w.SheetLayout(s, `{"op":"resize","index":1,"size":5000}`)
	w.SheetLayout(s, `{"op":"hideCol","index":0}`)
	w.SheetLayout(s, `{"op":"hideRow","index":2}`)

	sh := sheetOf(t, w)
	if len(sh.ColWidths) != 1 || sh.ColWidths[0].Size != 2000 {
		t.Fatalf("expected the width to be bounded, got %+v", sh.ColWidths)
	}

	if !slices.Equal(ints(sh.HiddenCols), []int{0}) || !slices.Equal(ints(sh.HiddenRows), []int{2, 9}) {
		t.Fatalf("unexpected hidden %v %v", sh.HiddenCols, sh.HiddenRows)
	}

	if l := layout.Get(); l.ColWidths[1] != 2000 || !slices.Equal(l.HiddenCols, []int{0}) {
		t.Fatalf("expected the layout in the bound state, got %+v", l)
	}

	w.SheetLayout(s, `{"op":"showAll"}`)
	sh = sheetOf(t, w)
	if len(sh.HiddenCols) != 0 || !slices.Equal(ints(sh.HiddenRows), []int{9}) {
		t.Fatalf("expected only the rows hidden by the application, got %v %v", sh.HiddenCols, sh.HiddenRows)
	}
}

func TestSheetNotCustomizable(t *testing.T) {
	m := &values{v: make([]int, 3)}
	w := openConfigured(t, m, func(s sheet.TSheet, _ core.Window) sheet.TSheet { return s.Customizable(false) })
	if sheetOf(t, w).Layout != 0 {
		t.Fatal("expected no layout binding")
	}
}

// Pasted values are applied like edits, cells without an editor are left out, and errors are shown per cell.
func TestSheetPaste(t *testing.T) {
	m := &values{v: make([]int, 10)}
	w := openConfigured(t, m, func(s sheet.TSheet, _ core.Window) sheet.TSheet { return s.AllowPaste(true) })
	s := w.Find(nagotest.Type[*proto.Sheet]())

	w.SheetPaste(s,
		sheet.Edit{Row: 1, Col: 0, Value: "7"},
		sheet.Edit{Row: 2, Col: 0, Value: "x"},
		sheet.Edit{Row: 1, Col: 1, Value: "99"}, // no editor
	)

	if m.v[1] != 7 {
		t.Fatalf("expected the pasted value, got %d", m.v[1])
	}

	if _, errText, _ := cellText(sheetOf(t, w), 2, 0); errText == "" {
		t.Fatal("expected the error of the invalid value")
	}

	if text, _, _ := cellText(sheetOf(t, w), 1, 1); text != "14" {
		t.Fatalf("expected the calculated cell, got %q", text)
	}
}

func TestSheetPasteIsOptIn(t *testing.T) {
	m := &values{v: make([]int, 3)}
	w := openConfigured(t, m, func(s sheet.TSheet, _ core.Window) sheet.TSheet { return s })
	if sheetOf(t, w).Paste != 0 {
		t.Fatal("expected pasting to be disabled by default")
	}
}

// A cell may show a view, e.g. a button, which works like anywhere else.
func TestSheetViewCell(t *testing.T) {
	clicked := 0
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.sheettest")
		cfg.RootView("sheet", func(wnd core.Window) core.View {
			return sheet.Sheet("buttons", 3, 1, func(row, col int) sheet.Cell {
				return sheet.Cell{Text: "Öffnen", View: ui.SecondaryButton(func() { clicked++ }).Title("Öffnen " + string(rune('A'+row)))}
			})
		})
	})

	w := app.Open(t, nil, "sheet")
	w.Click(w.Find(nagotest.Text("Öffnen B")))
	if clicked != 1 {
		t.Fatalf("expected the button in the cell to work, got %d clicks", clicked)
	}
}
