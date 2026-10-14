// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sheet

import (
	"encoding/json"
	"log/slog"
	"maps"
	"slices"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// TSheet is a basic component (Sheet).
// It shows a spreadsheet like table with frozen rows and columns, styled and merged cells, a keyboard navigation
// like in a spreadsheet application and the selection of ranges, which can be copied, e.g. into Excel.
//
// Only the cells of a window around the visible area are rendered: nago asks the cell function for exactly these
// cells, and the client requests the next window while the user scrolls. So a sheet with tens of thousands of rows
// is as fast as a small one, and a re-render after an edit, which changed anything, stays cheap.
//
// An edit is an event and passed to [TSheet.OnEdit], which may invoke any use case. The sheet re-renders
// afterwards and shows whatever the cell function returns then. If OnEdit returns an error, it is shown at the
// cell.
type TSheet struct {
	id         string
	rows, cols int
	cell       func(row, col int) Cell
	frozenRows int
	frozenCols int
	merges     []Merge
	hiddenRows map[int]bool
	hiddenCols map[int]bool
	density    Density
	colWidth   float64
	colWidths  map[int]float64
	gridColor  ui.Color
	onEdit     func(Edit) error
	rowHeights map[int]float64
	layout     *core.State[Layout]
	fixed      bool // the user may not change the layout
	paste      bool
	disabled   bool
	invisible  bool
	frame      ui.Frame
}

// Sheet creates a sheet of the given size. The id must be unique within the window, because the sheet keeps the
// window, which the user looks at, in states derived from it. The cell function is invoked only for the cells,
// which are rendered, and must be fast.
func Sheet(id string, rows, cols int, cell func(row, col int) Cell) TSheet {
	return TSheet{
		id:       id,
		rows:     max(rows, 0),
		cols:     max(cols, 0),
		cell:     cell,
		colWidth: 100,
		frame:    ui.Frame{Width: ui.Full, Height: ui.L480},
	}
}

// Frozen keeps the given number of leading rows and columns visible while scrolling, e.g. a header row.
func (c TSheet) Frozen(rows, cols int) TSheet {
	c.frozenRows = max(rows, 0)
	c.frozenCols = max(cols, 0)
	return c
}

// Merges joins rectangles of cells, see [Merge]. Merges must not overlap.
func (c TSheet) Merges(merges ...Merge) TSheet {
	c.merges = merges
	return c
}

// HiddenRows hides the rows with the given indices.
func (c TSheet) HiddenRows(rows ...int) TSheet {
	c.hiddenRows = set(rows)
	return c
}

// HiddenCols hides the columns with the given indices.
func (c TSheet) HiddenCols(cols ...int) TSheet {
	c.hiddenCols = set(cols)
	return c
}

// Density defines the height of the rows and the padding of the cells. The default is [DensityCompact]. On a
// touch device, the rows are at least as high as a finger needs.
func (c TSheet) Density(density Density) TSheet {
	c.density = density
	return c
}

// ColumnWidth sets the default width of the columns in dp, 100 by default.
func (c TSheet) ColumnWidth(width float64) TSheet {
	c.colWidth = width
	return c
}

// ColumnWidths sets the width of single columns in dp, by their index.
func (c TSheet) ColumnWidths(widths map[int]float64) TSheet {
	c.colWidths = widths
	return c
}

// RowHeights sets the height of single rows in dp, by their index. The other rows have the height of the
// [Density]. Text, which is wrapped, see [WrapWrap], shows as many lines as fit into the height.
func (c TSheet) RowHeights(heights map[int]float64) TSheet {
	c.rowHeights = heights
	return c
}

// Customizable lets the user change the width of the columns by dragging the edge of a header cell, i.e. a cell
// of a frozen row, and hide rows and columns by the context menu. It is enabled by default. See also
// [TSheet.Layout].
func (c TSheet) Customizable(customizable bool) TSheet {
	c.fixed = !customizable
	return c
}

// Layout binds the state, which keeps the changes of the user to the layout, e.g. to store them in the settings
// of the user. Without a bound state, the changes last as long as the window.
func (c TSheet) Layout(state *core.State[Layout]) TSheet {
	c.layout = state
	return c
}

// AllowPaste lets the user paste values, e.g. copied from Excel, into the cells, starting at the active one.
// Every value is passed to [TSheet.OnEdit] like an edit, and cells without an editor are left out. Pasting is
// disabled by default, because it invokes many use cases at once.
func (c TSheet) AllowPaste(allow bool) TSheet {
	c.paste = allow
	return c
}

// GridColor sets the color of the grid lines.
func (c TSheet) GridColor(color ui.Color) TSheet {
	c.gridColor = color
	return c
}

// OnEdit is invoked with every edit of a cell with an [Editor]. It may invoke any use case. An error is shown at
// the cell, until the cell is edited successfully.
func (c TSheet) OnEdit(fn func(Edit) error) TSheet {
	c.onEdit = fn
	return c
}

// Disabled shows the sheet without editors. Selecting and copying is still possible.
func (c TSheet) Disabled(disabled bool) TSheet {
	c.disabled = disabled
	return c
}

func (c TSheet) Visible(visible bool) TSheet {
	c.invisible = !visible
	return c
}

// Frame sets the size of the sheet. The sheet scrolls within it, the default is the full width and 480dp high.
func (c TSheet) Frame(frame ui.Frame) TSheet {
	c.frame = frame
	return c
}

func (c TSheet) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()
	prefix := "nago.sheet." + c.id

	viewport := core.StateOf[string](wnd, prefix+".viewport")
	editSeq := core.StateOf[int64](wnd, prefix+".editseq")
	errs := core.StateOf[map[string]string](wnd, prefix+".errors")

	edit := core.StateOf[string](wnd, prefix+".edit").Observe(func(raw string) {
		c.handleEdit(raw, editSeq, errs)
	})

	layout := c.layout
	if layout == nil {
		layout = core.StateOf[Layout](wnd, prefix+".layout")
	}

	// the cells depend on the layout of the user, e.g. its hidden columns
	c = c.withLayout(layout.Get())

	cp := core.StateOf[string](wnd, prefix+".copy").Observe(func(raw string) {
		c.handleCopy(wnd, raw)
	})

	var layoutPtr proto.Ptr
	if !c.fixed {
		layoutPtr = core.StateOf[string](wnd, prefix+".layoutop").Observe(func(raw string) {
			c.handleLayout(raw, layout)
		}).Ptr()
	}

	var pastePtr proto.Ptr
	if c.paste && !c.disabled && c.onEdit != nil {
		pastePtr = core.StateOf[string](wnd, prefix+".paste").Observe(func(raw string) {
			c.handlePaste(raw, editSeq, errs)
		}).Ptr()
	}

	win := c.window(viewport.Get())
	content := c.collect(ctx, win, errs.Get())

	return &proto.Sheet{
		Id:         proto.Str(c.id),
		RowCount:   proto.Int(c.rows),
		ColCount:   proto.Int(c.cols),
		FrozenRows: proto.Int(min(c.frozenRows, c.rows)),
		FrozenCols: proto.Int(min(c.frozenCols, c.cols)),
		RowHeight:  proto.Float(c.density.rowHeight()),
		ColWidth:   proto.Float(c.colWidth),
		ColWidths:  sizes(c.colWidths, c.cols),
		RowHeights: sizes(c.rowHeights, c.rows),
		HiddenRows: ints(c.hiddenRows),
		HiddenCols: ints(c.hiddenCols),
		WindowRow:  proto.Int(win.Row),
		WindowCol:  proto.Int(win.Col),
		WindowRows: proto.Int(win.Rows),
		WindowCols: proto.Int(win.Cols),
		Cells:      content.cells,
		Styles:     content.styles,
		Merges:     content.merges,
		Viewport:   viewport.Ptr(),
		Edit:       edit.Ptr(),
		EditSeq:    proto.Int(editSeq.Get()),
		Copy:       cp.Ptr(),
		GridColor:  proto.Color(c.gridColor),
		Frame:      frameOf(c.frame),
		Disabled:   proto.Bool(c.disabled),
		Invisible:  proto.Bool(c.invisible),
		Layout:     layoutPtr,
		Paste:      pastePtr,
	}
}

func sizes(m map[int]float64, count int) proto.SheetSizes {
	var res proto.SheetSizes
	for idx, size := range m {
		if idx >= 0 && idx < count && size > 0 {
			res = append(res, proto.SheetSize{Index: proto.Int(idx), Size: proto.Float(size)})
		}
	}

	slices.SortFunc(res, func(a, b proto.SheetSize) int { return int(a.Index - b.Index) })
	return res
}

// editRequest is written by the client into the edit state.
type editRequest struct {
	Seq   int64  `json:"seq"`
	Row   int    `json:"row"`
	Col   int    `json:"col"`
	Value string `json:"value"`
}

func (c TSheet) handleEdit(raw string, editSeq *core.State[int64], errs *core.State[map[string]string]) {
	var req editRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		return
	}

	// the sequence number tells the client, that its optimistic value has been processed, also if rejected
	if req.Seq > editSeq.Get() {
		editSeq.Set(req.Seq)
	}

	// the client is not trusted: only cells, which offer an editor right now, can be edited
	if c.disabled || c.onEdit == nil || req.Row < 0 || req.Row >= c.rows || req.Col < 0 || req.Col >= c.cols {
		return
	}

	updated := maps.Clone(errs.Get())
	if updated == nil {
		updated = map[string]string{}
	}

	c.apply(Edit{Row: req.Row, Col: req.Col, Value: req.Value}, updated)
	errs.Set(updated)
}

// apply passes the edit to OnEdit, if the cell has an editor, and remembers its error.
func (c TSheet) apply(e Edit, errs map[string]string) {
	if c.cell(e.Row, e.Col).Editor == EditorNone {
		return
	}

	key := cellKey(e.Row, e.Col)
	if err := c.onEdit(e); err != nil {
		errs[key] = errorText(err)
	} else {
		delete(errs, key)
	}
}

func (c TSheet) handleCopy(wnd core.Window, raw string) {
	r, ok := parseRange(raw)
	if !ok {
		return
	}

	text, err := c.tsv(clamp(r, c.rows, c.cols), wnd.Locale())
	if err != nil {
		slog.Warn("cannot copy the range of a sheet", "sheet", c.id, "err", err)
		return
	}

	if err := wnd.Clipboard().SetText(text); err != nil {
		slog.Warn("cannot write the clipboard", "sheet", c.id, "err", err)
	}
}

func set(values []int) map[int]bool {
	if len(values) == 0 {
		return nil
	}

	res := make(map[int]bool, len(values))
	for _, v := range values {
		res[v] = true
	}
	return res
}

func ints(set map[int]bool) proto.Ints {
	if len(set) == 0 {
		return nil
	}

	res := make(proto.Ints, 0, len(set))
	for v := range set {
		res = append(res, proto.Int(v))
	}

	sortInts(res)
	return res
}

func frameOf(f ui.Frame) proto.Frame {
	return proto.Frame{
		MinWidth:  proto.Length(f.MinWidth),
		MaxWidth:  proto.Length(f.MaxWidth),
		MinHeight: proto.Length(f.MinHeight),
		MaxHeight: proto.Length(f.MaxHeight),
		Width:     proto.Length(f.Width),
		Height:    proto.Length(f.Height),
		FlexProperties: proto.FlexProperties{
			Grow:          proto.Bool(f.FlexGrow),
			PreventShrink: proto.Bool(f.FlexPreventShrink),
		},
	}
}
