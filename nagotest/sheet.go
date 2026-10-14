// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui/sheet"
)

// sheetSeq numbers the edits of a sheet, like the frontend does.
var sheetSeq atomic.Int64

// SheetScroll requests the window of the single selected sheet, like the frontend does while the user scrolls.
// The window settles afterward.
func (w *Window) SheetScroll(s Selection, window sheet.Range) {
	w.t.Helper()

	start := time.Now()
	w.act(s, func(s Selection) proto.NagoEvent {
		sh := w.sheetOf(s)
		value := fmt.Sprintf("%d,%d,%d,%d", window.Row, window.Col, window.Rows, window.Cols)
		return &proto.UpdateStateValueRequested{StatePointer: sh.Viewport, Value: proto.Str(value), RID: w.nextRID()}
	})
	w.observe("sheet-scroll", s.matcher.desc, start)
}

// SheetEdit commits a value into a cell of the single selected sheet, like the frontend does after the user
// edited the cell. The window settles afterward.
func (w *Window) SheetEdit(s Selection, row, col int, value string) {
	w.t.Helper()

	start := time.Now()
	w.act(s, func(s Selection) proto.NagoEvent {
		sh := w.sheetOf(s)
		w.assertInteractive(s, s.Node())
		buf, err := json.Marshal(map[string]any{"seq": sheetSeq.Add(1), "row": row, "col": col, "value": value})
		if err != nil {
			w.t.Fatalf("nagotest: cannot encode sheet edit: %v", err)
		}
		return &proto.UpdateStateValueRequested{StatePointer: sh.Edit, Value: proto.Str(buf), RID: w.nextRID()}
	})
	w.observe("sheet-edit", s.matcher.desc, start)
}

func (w *Window) sheetOf(s Selection) *proto.Sheet {
	w.t.Helper()

	sh, ok := s.Node().Component.(*proto.Sheet)
	if !ok {
		w.t.Fatalf("nagotest: %s: %T is not a sheet", s.matcher, s.Node().Component)
	}

	return sh
}

// SheetLayout changes the layout of the single selected sheet, like the user does by dragging the edge of a
// header cell or by the context menu, e.g. {"op":"resize","index":2,"size":120}, {"op":"hideCol","index":2},
// {"op":"hideRow","index":5} or {"op":"showAll"}. The window settles afterward.
func (w *Window) SheetLayout(s Selection, op string) {
	w.t.Helper()

	start := time.Now()
	w.act(s, func(s Selection) proto.NagoEvent {
		sh := w.sheetOf(s)
		if sh.Layout == 0 {
			w.t.Fatalf("nagotest: %s: the sheet is not customizable", s.matcher)
		}
		return &proto.UpdateStateValueRequested{StatePointer: sh.Layout, Value: proto.Str(op), RID: w.nextRID()}
	})
	w.observe("sheet-layout", s.matcher.desc, start)
}

// SheetPaste pastes values into the single selected sheet, like the frontend does after the user pasted, e.g.
// from Excel. The window settles afterward.
func (w *Window) SheetPaste(s Selection, edits ...sheet.Edit) {
	w.t.Helper()

	start := time.Now()
	w.act(s, func(s Selection) proto.NagoEvent {
		sh := w.sheetOf(s)
		if sh.Paste == 0 {
			w.t.Fatalf("nagotest: %s: the sheet does not allow pasting", s.matcher)
		}

		type edit struct {
			Row   int    `json:"row"`
			Col   int    `json:"col"`
			Value string `json:"value"`
		}
		req := struct {
			Seq   int64  `json:"seq"`
			Edits []edit `json:"edits"`
		}{Seq: sheetSeq.Add(1)}
		for _, e := range edits {
			req.Edits = append(req.Edits, edit{Row: e.Row, Col: e.Col, Value: e.Value})
		}

		buf, err := json.Marshal(req)
		if err != nil {
			w.t.Fatalf("nagotest: cannot encode sheet paste: %v", err)
		}
		return &proto.UpdateStateValueRequested{StatePointer: sh.Paste, Value: proto.Str(buf), RID: w.nextRID()}
	})
	w.observe("sheet-paste", s.matcher.desc, start)
}

// SheetSort sorts the single selected sheet by a column, like the user does with the arrow in the header. The
// window settles afterward.
func (w *Window) SheetSort(s Selection, sorting sheet.Sorting) {
	w.t.Helper()

	start := time.Now()
	w.act(s, func(s Selection) proto.NagoEvent {
		sh := w.sheetOf(s)
		if sh.Header.Sort == 0 {
			w.t.Fatalf("nagotest: %s: the sheet offers no sorting", s.matcher)
		}
		value := fmt.Sprintf("%d,%d", sorting.Col, sorting.Dir)
		return &proto.UpdateStateValueRequested{StatePointer: sh.Header.Sort, Value: proto.Str(value), RID: w.nextRID()}
	})
	w.observe("sheet-sort", s.matcher.desc, start)
}

// SheetFilter filters the single selected sheet by the text of a column, like the user does in the header. An
// empty text removes the filter. The window settles afterward.
func (w *Window) SheetFilter(s Selection, col int, text string) {
	w.t.Helper()

	start := time.Now()
	w.act(s, func(s Selection) proto.NagoEvent {
		sh := w.sheetOf(s)
		if sh.Header.Filter == 0 {
			w.t.Fatalf("nagotest: %s: the sheet offers no filtering", s.matcher)
		}
		buf, err := json.Marshal(map[string]any{"col": col, "text": text})
		if err != nil {
			w.t.Fatalf("nagotest: cannot encode sheet filter: %v", err)
		}
		return &proto.UpdateStateValueRequested{StatePointer: sh.Header.Filter, Value: proto.Str(buf), RID: w.nextRID()}
	})
	w.observe("sheet-filter", s.matcher.desc, start)
}
