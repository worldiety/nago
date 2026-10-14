// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sheet

import (
	"fmt"
	"strconv"
	"testing"

	"go.wdy.de/nago/presentation/proto"
	"golang.org/x/text/language"
)

var calls int

func numbers(row, col int) Cell {
	calls++
	return Cell{Text: fmt.Sprintf("%d/%d", row, col), Value: strconv.Itoa(row*1000+col) + ".5", Type: ValueNumber}
}

func positions(cells proto.SheetCells) map[string]bool {
	res := map[string]bool{}
	for _, c := range cells {
		res[cellKey(int(c.Row), int(c.Col))] = true
	}
	return res
}

func TestWindow(t *testing.T) {
	s := Sheet("s", 50_000, 30, numbers).Frozen(1, 1)

	if got := s.window(""); got != (Range{Row: 1, Col: 1, Rows: defaultWindowRows, Cols: 29}) {
		t.Fatalf("unexpected default window %+v", got)
	}

	if got := s.window("49990,0,100,10"); got != (Range{Row: 49990, Col: 0, Rows: 10, Cols: 10}) {
		t.Fatalf("expected the window to be clamped to the sheet, got %+v", got)
	}

	if got := s.window("0,0,100000,30"); got.Rows*got.Cols > maxWindowCells {
		t.Fatalf("expected the window to be bounded, got %+v", got)
	}

	if got := s.window("x,1,2,3"); got.Rows != defaultWindowRows {
		t.Fatalf("expected an invalid window to fall back, got %+v", got)
	}
}

// Only the frozen rows and columns and the window are rendered, so a large sheet costs as much as a small one.
func TestCollectRendersOnlyTheWindow(t *testing.T) {
	calls = 0
	s := Sheet("s", 50_000, 30, numbers).Frozen(1, 1)
	got := positions(s.collect(Range{Row: 1000, Col: 10, Rows: 3, Cols: 2}, nil).cells)

	// corner, frozen row, frozen column and window
	for _, want := range []string{"0:0", "0:10", "0:11", "1000:0", "1002:0", "1000:10", "1002:11"} {
		if !got[want] {
			t.Errorf("expected cell %s", want)
		}
	}

	if len(got) != 4*3 || calls != 12 {
		t.Fatalf("expected 12 cells, got %d cells and %d calls", len(got), calls)
	}
}

func TestCollectHiddenAndMerged(t *testing.T) {
	s := Sheet("s", 100, 10, numbers).
		HiddenRows(11).
		HiddenCols(3).
		Merges(Merge{Row: 5, Col: 1, Rows: 6, Cols: 2})

	res := s.collect(Range{Row: 10, Col: 0, Rows: 3, Cols: 5}, nil)
	got := positions(res.cells)

	if got["11:0"] || got["10:3"] {
		t.Error("hidden cells must be omitted")
	}

	if got["10:1"] || got["10:2"] {
		t.Error("cells covered by a merge must be omitted")
	}

	if !got["5:1"] {
		t.Error("the top left cell of a merge reaching into the window must be included")
	}

	if len(res.merges) != 1 || res.merges[0].Rows != 6 {
		t.Fatalf("expected the merge, got %+v", res.merges)
	}
}

func TestCollectDeduplicatesStyles(t *testing.T) {
	bold := Style{Bold: true}
	s := Sheet("s", 10, 10, func(row, col int) Cell {
		if row%2 == 0 {
			return Cell{Text: "x", Style: bold}
		}
		return Cell{Text: "y"}
	})

	res := s.collect(Range{Rows: 10, Cols: 10}, nil)
	if len(res.styles) != 1 {
		t.Fatalf("expected one style, got %d", len(res.styles))
	}

	for _, c := range res.cells {
		if (int(c.Row)%2 == 0) != (c.Style == 1) {
			t.Fatalf("unexpected style reference %+v", c)
		}
	}
}

func TestTSV(t *testing.T) {
	s := Sheet("s", 3, 3, func(row, col int) Cell {
		switch col {
		case 0:
			return Cell{Text: "1.234,50 €", Value: "1234.5", Type: ValueNumber}
		case 1:
			return Cell{Text: "a\tb"}
		default:
			return Cell{Text: "x"}
		}
	}).HiddenRows(1)

	got, err := s.tsv(Range{Rows: 3, Cols: 3}, language.German)
	if err != nil {
		t.Fatal(err)
	}

	want := "1234,5\t\"a\tb\"\tx\r\n1234,5\t\"a\tb\"\tx"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	got, _ = s.tsv(Range{Rows: 1, Cols: 1}, language.English)
	if got != "1234.5" {
		t.Fatalf("expected a decimal point for English, got %q", got)
	}

	if _, err := Sheet("s", 1_000_000, 10, numbers).tsv(Range{Rows: 1_000_000, Cols: 10}, language.German); err == nil {
		t.Fatal("expected a too large range to be rejected")
	}
}

func BenchmarkCollect(b *testing.B) {
	s := Sheet("s", 50_000, 30, numbers).Frozen(1, 1)
	win := s.window("25000,1,80,29")
	b.ReportAllocs()
	for b.Loop() {
		s.collect(win, nil)
	}
}
