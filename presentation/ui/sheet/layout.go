// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sheet

import (
	"encoding/json"
	"maps"
	"slices"
	"strconv"
	"strings"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

const (
	// minColWidth and maxColWidth bound the width, which a user may give a column.
	minColWidth = 24
	maxColWidth = 2000

	// maxPasteCells bounds the values of a single paste, because each of them invokes a use case.
	maxPasteCells = 10_000
)

// withLayout returns the sheet with the changes of the user applied: widths of the user win over the ones of the
// application, and hidden rows and columns of both are hidden.
func (c TSheet) withLayout(l Layout) TSheet {
	if len(l.ColWidths) > 0 {
		widths := maps.Clone(c.colWidths)
		if widths == nil {
			widths = map[int]float64{}
		}
		maps.Copy(widths, l.ColWidths)
		c.colWidths = widths
	}

	c.hiddenRows = union(c.hiddenRows, l.HiddenRows)
	c.hiddenCols = union(c.hiddenCols, l.HiddenCols)
	return c
}

func union(set map[int]bool, values []int) map[int]bool {
	if len(values) == 0 {
		return set
	}

	res := maps.Clone(set)
	if res == nil {
		res = map[int]bool{}
	}

	for _, v := range values {
		res[v] = true
	}
	return res
}

// userHidden returns the rows and columns, which only the user has hidden, so that the user may show them again.
func (c TSheet) userHidden(l Layout) proto.SheetUserHidden {
	var res proto.SheetUserHidden
	for _, row := range l.HiddenRows {
		if row >= 0 && row < c.rows && !c.hiddenRows[row] {
			res.Rows = append(res.Rows, proto.Int(row))
		}
	}

	for _, col := range l.HiddenCols {
		if col >= 0 && col < c.cols && !c.hiddenCols[col] {
			res.Cols = append(res.Cols, proto.Int(col))
		}
	}

	slices.Sort(res.Rows)
	slices.Sort(res.Cols)
	return res
}

func (c TSheet) handleSort(raw string) {
	parts := strings.Split(raw, ",")
	if len(parts) != 2 || c.onSort == nil {
		return
	}

	col, err1 := strconv.Atoi(parts[0])
	dir, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || col < 0 || col >= c.cols || SortDir(dir) < SortNone || SortDir(dir) > SortDesc {
		return
	}

	if len(c.sortable) > 0 && !slices.Contains(c.sortable, col) {
		return
	}

	c.onSort(Sorting{Col: col, Dir: SortDir(dir)})
}

// filterRequest is written by the client into the filter state.
type filterRequest struct {
	Col  int    `json:"col"`
	Text string `json:"text"`
}

func (c TSheet) handleFilter(raw string) {
	var req filterRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil || c.onFilter == nil || req.Col < 0 || req.Col >= c.cols {
		return
	}

	if len(c.filterable) > 0 && !slices.Contains(c.filterable, req.Col) {
		return
	}

	c.onFilter(req.Col, strings.TrimSpace(req.Text))
}

// layoutOp is written by the client into the layout state.
type layoutOp struct {
	Op    string  `json:"op"`
	Index int     `json:"index"`
	Size  float64 `json:"size"`
}

func (c TSheet) handleLayout(raw string, state *core.State[Layout]) {
	var op layoutOp
	if err := json.Unmarshal([]byte(raw), &op); err != nil {
		return
	}

	l := state.Get()
	switch op.Op {
	case "resize":
		if op.Index < 0 || op.Index >= c.cols {
			return
		}
		l.ColWidths = maps.Clone(l.ColWidths)
		if l.ColWidths == nil {
			l.ColWidths = map[int]float64{}
		}
		l.ColWidths[op.Index] = min(max(op.Size, minColWidth), maxColWidth)
	case "hideCol":
		if op.Index < 0 || op.Index >= c.cols || slices.Contains(l.HiddenCols, op.Index) {
			return
		}
		l.HiddenCols = append(slices.Clone(l.HiddenCols), op.Index)
	case "hideRow":
		if op.Index < 0 || op.Index >= c.rows || slices.Contains(l.HiddenRows, op.Index) {
			return
		}
		l.HiddenRows = append(slices.Clone(l.HiddenRows), op.Index)
	case "showAll":
		l.HiddenRows = nil
		l.HiddenCols = nil
	case "showCol":
		l.HiddenCols = slices.DeleteFunc(slices.Clone(l.HiddenCols), func(i int) bool { return i == op.Index })
	case "showRow":
		l.HiddenRows = slices.DeleteFunc(slices.Clone(l.HiddenRows), func(i int) bool { return i == op.Index })
	default:
		return
	}

	state.Set(l)
	state.Notify()
}

// pasteRequest is written by the client into the paste state.
type pasteRequest struct {
	Seq   int64         `json:"seq"`
	Edits []editRequest `json:"edits"`
}

func (c TSheet) handlePaste(raw string, editSeq *core.State[int64], errs *core.State[map[string]string]) {
	var req pasteRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		return
	}

	if req.Seq > editSeq.Get() {
		editSeq.Set(req.Seq)
	}

	if len(req.Edits) > maxPasteCells {
		req.Edits = req.Edits[:maxPasteCells]
	}

	updated := maps.Clone(errs.Get())
	if updated == nil {
		updated = map[string]string{}
	}

	for _, e := range req.Edits {
		if e.Row < 0 || e.Row >= c.rows || e.Col < 0 || e.Col >= c.cols || c.hiddenRows[e.Row] || c.hiddenCols[e.Col] {
			continue
		}

		c.apply(Edit{Row: e.Row, Col: e.Col, Value: e.Value}, updated)
	}

	errs.Set(updated)
}
