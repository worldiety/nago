// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sheet_test

import (
	"strconv"
	"sync"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui/sheet"
)

// values is the model of the application: one editable number per row and its double as a calculated metric.
type values struct {
	mutex sync.Mutex
	v     []int
	calls int
}

func (m *values) cell(row, col int) sheet.Cell {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.calls++
	if col == 0 {
		return sheet.Cell{Text: strconv.Itoa(m.v[row]), Type: sheet.ValueNumber, Editor: sheet.EditorNumber}
	}
	return sheet.Cell{Text: strconv.Itoa(m.v[row] * 2), Type: sheet.ValueNumber}
}

func (m *values) edit(e sheet.Edit) error {
	n, err := strconv.Atoi(e.Value)
	if err != nil {
		return std.NewLocalizedError("Ungültig", "Bitte eine Zahl eingeben.")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.v[e.Row] = n
	return nil
}

func openSheet(t *testing.T, m *values) *nagotest.Window {
	t.Helper()
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.sheettest")
		cfg.RootView("sheet", func(wnd core.Window) core.View {
			return sheet.Sheet("values", len(m.v), 2, m.cell).OnEdit(m.edit)
		})
	})

	return app.Open(t, nil, "sheet")
}

func sheetOf(t *testing.T, w *nagotest.Window) *proto.Sheet {
	t.Helper()
	return w.Find(nagotest.Type[*proto.Sheet]()).Node().Component.(*proto.Sheet)
}

func cellText(sh *proto.Sheet, row, col int) (string, string, bool) {
	for _, c := range sh.Cells {
		if int(c.Row) == row && int(c.Col) == col {
			return string(c.Text), string(c.Error), true
		}
	}
	return "", "", false
}

// A large sheet renders only its window, and scrolling renders the next one.
func TestSheetRendersTheRequestedWindow(t *testing.T) {
	m := &values{v: make([]int, 100_000)}
	w := openSheet(t, m)

	if sh := sheetOf(t, w); len(sh.Cells) > 200 || int(sh.RowCount) != 100_000 {
		t.Fatalf("expected only the default window of %d cells for %d rows", len(sh.Cells), sh.RowCount)
	}

	w.SheetScroll(w.Find(nagotest.Type[*proto.Sheet]()), sheet.Range{Row: 90_000, Col: 0, Rows: 50, Cols: 2})
	sh := sheetOf(t, w)
	if int(sh.WindowRow) != 90_000 || len(sh.Cells) != 100 {
		t.Fatalf("expected the requested window, got row %d with %d cells", sh.WindowRow, len(sh.Cells))
	}

	if _, _, ok := cellText(sh, 90_049, 1); !ok {
		t.Fatal("expected the last cell of the window")
	}
}

// An edit invokes the use case, the sheet shows the recalculated cells, and a rejected edit shows the error.
func TestSheetEdit(t *testing.T) {
	m := &values{v: make([]int, 10)}
	w := openSheet(t, m)
	s := w.Find(nagotest.Type[*proto.Sheet]())

	w.SheetEdit(s, 3, 0, "21")
	if text, _, _ := cellText(sheetOf(t, w), 3, 1); text != "42" {
		t.Fatalf("expected the recalculated metric, got %q", text)
	}

	if seq := sheetOf(t, w).EditSeq; seq == 0 {
		t.Fatal("expected the processed edit to be acknowledged")
	}

	w.SheetEdit(s, 3, 0, "abc")
	if text, errText, _ := cellText(sheetOf(t, w), 3, 0); text != "21" || errText != "Bitte eine Zahl eingeben." {
		t.Fatalf("expected the old value with the error, got %q, %q", text, errText)
	}

	w.SheetEdit(s, 3, 0, "5")
	if _, errText, _ := cellText(sheetOf(t, w), 3, 0); errText != "" {
		t.Fatalf("expected the error to be gone after a valid edit, got %q", errText)
	}

	// a cell without an editor cannot be edited, whatever the client sends
	w.SheetEdit(s, 3, 1, "999")
	if m.v[3] != 5 {
		t.Fatalf("expected a read-only cell to be ignored, got %d", m.v[3])
	}
}
