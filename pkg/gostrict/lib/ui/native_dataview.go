// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//go:build !gostrict

package ui

import (
	"maps"
	"slices"
	"strconv"

	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/dataview"
)

// The data view of the program is a specification, which builds the data view of nago, when it renders, so that its
// options accumulate like the ones of nago. The rows are the indices 0 to rows-1, see dataview.FromSlice.

type dataView struct {
	wnd                                   core.Window
	rows                                  int32
	fields                                []TDataViewField
	action                                func(row int32)
	nextIndicator, hideSelection, search  bool
	create                                func()
	createOptions                         []TDataViewCreateOption
	createView, actionBar                 core.View
	selectOptions                         []TDataViewSelectOption
	selectionChanged                      func(rows list.List[int32])
	style                                 DataViewStyle
	cardTitle, listTitle, listDescription string
	cardHints, listHints                  map[string]DataViewFormatHint
	highlight                             []int32
	widths                                map[string]Length
}

type dataViewField struct {
	id, name   string
	view       func(row int32) View
	cmp        func(a, b int32) int32
	minSize    WindowSizeClass
	hasMinSize bool
}

type dataViewSelect struct {
	icon                         SVG
	name                         string
	delete                       bool
	wnd                          core.Window
	action                       func(rows list.List[int32]) error
	visible                      func(rows list.List[int32]) bool
	confirmTitle, confirmMessage string
}

type dataViewCreate struct {
	icon   SVG
	name   string
	action func() error
}

func newDataView(wnd core.Window, rows int32, fields []TDataViewField) dataView {
	return dataView{wnd: wnd, rows: max(rows, 0), fields: fields, search: true}
}

func (c TDataView) with(fn func(d *dataView)) TDataView {
	d := get[dataView](c.ref)
	fn(&d)
	return TDataView{ref: wrap(d)}
}

func (f TDataViewField) with(fn func(d *dataViewField)) TDataViewField {
	d := get[dataViewField](f.ref)
	fn(&d)
	return TDataViewField{ref: wrap(d)}
}

func (o TDataViewSelectOption) with(fn func(d *dataViewSelect)) TDataViewSelectOption {
	d := get[dataViewSelect](o.ref)
	fn(&d)
	return TDataViewSelectOption{ref: wrap(d)}
}

func withHint(m map[string]DataViewFormatHint, field string, hint DataViewFormatHint) map[string]DataViewFormatHint {
	m = maps.Clone(m)
	if m == nil {
		m = map[string]DataViewFormatHint{}
	}
	m[field] = hint
	return m
}

func withWidth(m map[string]Length, field string, width Length) map[string]Length {
	m = maps.Clone(m)
	if m == nil {
		m = map[string]Length{}
	}
	m[field] = width
	return m
}

type row = dataview.Element[int]

// rowsOf returns the indices of the identities of the selected rows.
func rowsOf(ids []dataview.Idx) list.List[int32] {
	res := list.New[int32]()
	for _, id := range ids {
		if i, ok := id.Int(); ok {
			res.Append(int32(i))
		}
	}
	return res
}

// Render builds the data view of nago.
func (d dataView) Render(ctx core.RenderContext) core.RenderNode {
	return d.nago().Render(ctx)
}

func (d dataView) nago() dataview.TDataView[row, dataview.Idx] {
	indices := make([]int, d.rows)
	for i := range indices {
		indices[i] = i
	}

	var fields []dataview.Field[row]
	for _, h := range d.fields {
		f := get[dataViewField](h.ref)
		field := dataview.Field[row]{ID: dataview.FieldID(f.id), Name: f.name}
		if f.view != nil {
			view := f.view
			field.Map = func(r row) core.View { return viewOf(view(int32(r.Value))) }
		}
		if f.cmp != nil {
			cmp := f.cmp
			field.Comparator = func(a, b row) int { return int(cmp(int32(a.Value), int32(b.Value))) }
		}
		if f.hasMinSize {
			class := core.WindowSizeClass(f.minSize)
			field.Visible = func(ctx dataview.FieldContext) bool { return ctx.Window.Info().SizeClass >= class }
		}
		fields = append(fields, field)
	}

	v := dataview.FromSlice(d.wnd, indices, fields).
		NextActionIndicator(d.nextIndicator).
		Selection(!d.hideSelection).
		Search(d.search).
		Style(dataview.Style(d.style))
	if d.action != nil {
		action := d.action
		v = v.Action(func(r row) { action(int32(r.Value)) })
	}
	if d.create != nil {
		v = v.CreateAction(d.create)
	}
	if len(d.createOptions) > 0 {
		var opts []dataview.CreateOption
		for _, h := range d.createOptions {
			o := get[dataViewCreate](h.ref)
			opts = append(opts, dataview.CreateOption{Icon: core.SVG(o.icon), Name: o.name, Action: o.action})
		}
		v = v.CreateOptions(opts...)
	}
	if d.createView != nil {
		v = v.CreateActionView(d.createView)
	}
	if d.actionBar != nil {
		v = v.ActionBarContent(d.actionBar)
	}
	if len(d.selectOptions) > 0 {
		var opts []dataview.SelectOption[dataview.Idx]
		for _, h := range d.selectOptions {
			opts = append(opts, get[dataViewSelect](h.ref).nago())
		}
		v = v.SelectOptions(opts...)
	}
	if d.selectionChanged != nil {
		changed := d.selectionChanged
		v = v.SelectionChanged(func(ids []dataview.Idx) { changed(rowsOf(ids)) })
	}

	card := dataview.CardOptions{Title: dataview.FieldID(d.cardTitle), Hints: map[dataview.FieldID]dataview.FormatHint{}}
	for k, h := range d.cardHints {
		card.Hints[dataview.FieldID(k)] = dataview.FormatHint(h)
	}
	lst := dataview.ListOptions[dataview.Idx]{Title: dataview.FieldID(d.listTitle), Description: dataview.FieldID(d.listDescription), Hints: map[dataview.FieldID]dataview.FormatHint{}, Highlight: map[dataview.Idx]bool{}}
	for k, h := range d.listHints {
		lst.Hints[dataview.FieldID(k)] = dataview.FormatHint(h)
	}
	for _, r := range d.highlight {
		lst.Highlight[dataview.Idx(strconv.Itoa(int(r)))] = true
	}
	tbl := dataview.TableOptions{ColumnWidths: map[dataview.FieldID]nui.Length{}}
	for _, k := range slices.Sorted(maps.Keys(d.widths)) {
		tbl.ColumnWidths[dataview.FieldID(k)] = nui.Length(d.widths[k])
	}
	return v.CardOptions(card).ListOptions(lst).TableOptions(tbl)
}

func (o dataViewSelect) nago() dataview.SelectOption[dataview.Idx] {
	action := func(ids []dataview.Idx) error { return o.action(rowsOf(ids)) }
	var res dataview.SelectOption[dataview.Idx]
	if o.delete {
		res = dataview.NewSelectOptionDelete(o.wnd, action)
	} else {
		res = dataview.SelectOption[dataview.Idx]{Icon: core.SVG(o.icon), Name: o.name, Action: action}
	}
	if o.visible != nil {
		visible := o.visible
		res.Visible = func(ids []dataview.Idx) bool { return visible(rowsOf(ids)) }
	}
	if o.confirmTitle != "" || o.confirmMessage != "" {
		res.ConfirmDialog = func([]dataview.Idx) dataview.ConfirmDialog[dataview.Idx] {
			return dataview.ConfirmDialog[dataview.Idx]{Title: o.confirmTitle, Message: o.confirmMessage}
		}
	}
	return res
}
