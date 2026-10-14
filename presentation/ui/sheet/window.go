// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sheet

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"go.wdy.de/nago/presentation/proto"
)

const (
	// defaultWindowRows and defaultWindowCols are rendered, before the client has reported its viewport.
	defaultWindowRows = 80
	defaultWindowCols = 30

	// maxWindowCells bounds the window, which a client may request, so that a client cannot let the server
	// render the whole sheet at once.
	maxWindowCells = 20_000
)

// interval is a half open range of indices.
type interval struct{ from, to int }

func (i interval) contains(v int) bool { return v >= i.from && v < i.to }

func (i interval) intersects(from, to int) bool { return from < i.to && to > i.from }

func parseRange(s string) (Range, bool) {
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return Range{}, false
	}

	var v [4]int
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 {
			return Range{}, false
		}
		v[i] = n
	}

	return Range{Row: v[0], Col: v[1], Rows: v[2], Cols: v[3]}, true
}

func formatRange(r Range) string {
	return fmt.Sprintf("%d,%d,%d,%d", r.Row, r.Col, r.Rows, r.Cols)
}

// clamp keeps the range within the sheet.
func clamp(r Range, rows, cols int) Range {
	r.Row = min(max(r.Row, 0), rows)
	r.Col = min(max(r.Col, 0), cols)
	r.Rows = min(max(r.Rows, 0), rows-r.Row)
	r.Cols = min(max(r.Cols, 0), cols-r.Col)
	return r
}

// window returns the window to render: the requested one within the sheet and the budget, or a default one.
func (c TSheet) window(requested string) Range {
	r, ok := parseRange(requested)
	if !ok {
		r = Range{Row: c.frozenRows, Col: c.frozenCols, Rows: defaultWindowRows, Cols: defaultWindowCols}
	}

	if r.Rows*r.Cols > maxWindowCells {
		// keep the requested columns, they are usually few, and shorten the rows
		r.Cols = min(r.Cols, maxWindowCells)
		r.Rows = maxWindowCells / max(r.Cols, 1)
	}

	return clamp(r, c.rows, c.cols)
}

// content is what a render of a window transfers.
type content struct {
	cells  proto.SheetCells
	styles proto.SheetStyles
	merges proto.SheetMerges
}

// collect renders the cells of the frozen rows and columns and of the window. Hidden cells and cells covered by
// a merge are omitted, but the top left cell of a merge is included, even if it lies outside.
func (c TSheet) collect(win Range, errs map[string]string) content {
	rowAreas := []interval{{0, min(c.frozenRows, c.rows)}, {win.Row, win.Row + win.Rows}}
	colAreas := []interval{{0, min(c.frozenCols, c.cols)}, {win.Col, win.Col + win.Cols}}

	rows := indices(rowAreas, c.hiddenRows)
	cols := indices(colAreas, c.hiddenCols)

	var res content
	styleIdx := map[Style]int{}

	var merges []Merge
	for _, m := range c.merges {
		if m.Rows <= 0 || m.Cols <= 0 || (m.Rows == 1 && m.Cols == 1) {
			continue
		}

		if intersectsAny(rowAreas, m.Row, m.Row+m.Rows) && intersectsAny(colAreas, m.Col, m.Col+m.Cols) {
			merges = append(merges, m)
			res.merges = append(res.merges, proto.SheetMerge{Row: proto.Int(m.Row), Col: proto.Int(m.Col), Rows: proto.Int(m.Rows), Cols: proto.Int(m.Cols)})
		}
	}

	covered := func(row, col int) bool {
		for _, m := range merges {
			if m.contains(row, col) && (row != m.Row || col != m.Col) {
				return true
			}
		}
		return false
	}

	add := func(row, col int) {
		cell := c.cell(row, col)
		pc := proto.SheetCell{
			Row:     proto.Int(row),
			Col:     proto.Int(col),
			Text:    proto.Str(cell.Text),
			Value:   proto.Str(cell.Value),
			Type:    proto.SheetValueType(cell.Type),
			Editor:  proto.SheetEditor(cell.Editor),
			Options: strs(cell.Options),
			Error:   proto.Str(errs[cellKey(row, col)]),
		}

		if c.disabled {
			pc.Editor = proto.SheetEditorNone
		}

		if cell.Style != (Style{}) {
			idx, ok := styleIdx[cell.Style]
			if !ok {
				res.styles = append(res.styles, styleOf(cell.Style))
				idx = len(res.styles)
				styleIdx[cell.Style] = idx
			}
			pc.Style = proto.Int(idx)
		}

		res.cells = append(res.cells, pc)
	}

	inRows := map[int]bool{}
	for _, r := range rows {
		inRows[r] = true
	}

	inCols := map[int]bool{}
	for _, col := range cols {
		inCols[col] = true
	}

	for _, row := range rows {
		for _, col := range cols {
			if !covered(row, col) {
				add(row, col)
			}
		}
	}

	// the top left cell of a merge, which reaches into the window from outside
	for _, m := range merges {
		if (inRows[m.Row] && inCols[m.Col]) || c.hiddenRows[m.Row] || c.hiddenCols[m.Col] {
			continue
		}
		add(m.Row, m.Col)
	}

	return res
}

func intersectsAny(areas []interval, from, to int) bool {
	for _, a := range areas {
		if a.intersects(from, to) {
			return true
		}
	}
	return false
}

// indices returns the sorted, distinct and visible indices of the areas.
func indices(areas []interval, hidden map[int]bool) []int {
	var res []int
	seen := map[int]bool{}
	for _, a := range areas {
		for i := a.from; i < a.to; i++ {
			if seen[i] || hidden[i] {
				continue
			}
			seen[i] = true
			res = append(res, i)
		}
	}

	slices.Sort(res)
	return res
}

func cellKey(row, col int) string {
	return strconv.Itoa(row) + ":" + strconv.Itoa(col)
}

func strs(values []string) proto.Strings {
	if len(values) == 0 {
		return nil
	}

	res := make(proto.Strings, 0, len(values))
	for _, v := range values {
		res = append(res, proto.Str(v))
	}
	return res
}

func styleOf(s Style) proto.SheetStyle {
	return proto.SheetStyle{
		BackgroundColor: proto.Color(s.BackgroundColor),
		Color:           proto.Color(s.Color),
		Bold:            proto.Bool(s.Bold),
		Italic:          proto.Bool(s.Italic),
		Underline:       proto.Bool(s.Underline),
		Align:           proto.SheetAlign(s.Align),
		Wrap:            proto.SheetWrap(s.Wrap),
		BorderTop:       borderOf(s.Borders.Top),
		BorderRight:     borderOf(s.Borders.Right),
		BorderBottom:    borderOf(s.Borders.Bottom),
		BorderLeft:      borderOf(s.Borders.Left),
		PaddingX:        proto.Float(s.PaddingX),
		FontSize:        proto.Float(s.FontSize),
	}
}

func borderOf(l Line) proto.SheetBorder {
	return proto.SheetBorder{
		Color: proto.Color(l.Color),
		Width: proto.Float(l.Width),
		Style: proto.SheetLineStyle(l.Style),
	}
}
