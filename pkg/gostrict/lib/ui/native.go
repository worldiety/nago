// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui/sheet"
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	nsheet "go.wdy.de/nago/presentation/ui/sheet"
)

// Of returns the handle of a view of nago.
func Of(v core.View) View {
	return wrap(v)
}

// Core returns the view of nago or nil, e.g. for the zero handle, which a failed callback returns.
func (v View) Core() core.View {
	cv, _ := v.ref.Value().(core.View)
	return cv
}

// HostRef returns the reference of the handle, so that it crosses the boundary of an engine.
func (v View) HostRef() js.Ref {
	return v.ref
}

// HostFromRef returns the handle of a reference.
func (View) HostFromRef(ref js.Ref) any {
	return View{ref: ref}
}

func cellOf(c sheet.Cell) nsheet.Cell {
	return nsheet.Cell{
		Text:   c.Text,
		Value:  c.Value,
		Type:   valueTypeOf(c.Type),
		Style:  styleOf(c.Style),
		Editor: editorOf(c.Editor),
	}
}

func styleOf(s sheet.Style) nsheet.Style {
	return nsheet.Style{
		BackgroundColor: ui.Color(s.BackgroundColor),
		Color:           ui.Color(s.Color),
		Bold:            s.Bold,
		Italic:          s.Italic,
		Underline:       s.Underline,
		Align:           alignOf(s.Align),
		Wrap:            wrapOf(s.Wrap),
		Borders: nsheet.Borders{
			Top:    lineOf(s.Borders.Top),
			Right:  lineOf(s.Borders.Right),
			Bottom: lineOf(s.Borders.Bottom),
			Left:   lineOf(s.Borders.Left),
		},
		PaddingX: s.PaddingX,
		FontSize: s.FontSize,
	}
}

func lineOf(l sheet.Line) nsheet.Line {
	return nsheet.Line{Color: ui.Color(l.Color), Width: l.Width, Style: lineStyleOf(l.Style)}
}

func valueTypeOf(t sheet.ValueType) nsheet.ValueType {
	switch t {
	case sheet.ValueNumber:
		return nsheet.ValueNumber
	case sheet.ValueBool:
		return nsheet.ValueBool
	case sheet.ValueDate:
		return nsheet.ValueDate
	}
	return nsheet.ValueText
}

func editorOf(e sheet.Editor) nsheet.Editor {
	switch e {
	case sheet.EditorText:
		return nsheet.EditorText
	case sheet.EditorNumber:
		return nsheet.EditorNumber
	case sheet.EditorCheckbox:
		return nsheet.EditorCheckbox
	}
	return nsheet.EditorNone
}

func alignOf(a sheet.Align) nsheet.Align {
	switch a {
	case sheet.AlignStart:
		return nsheet.AlignStart
	case sheet.AlignCenter:
		return nsheet.AlignCenter
	case sheet.AlignEnd:
		return nsheet.AlignEnd
	}
	return nsheet.AlignAuto
}

func wrapOf(w sheet.Wrap) nsheet.Wrap {
	switch w {
	case sheet.WrapEllipsis:
		return nsheet.WrapEllipsis
	case sheet.WrapWrap:
		return nsheet.WrapWrap
	}
	return nsheet.WrapClip
}

func lineStyleOf(s sheet.LineStyle) nsheet.LineStyle {
	switch s {
	case sheet.LineSolid:
		return nsheet.LineSolid
	case sheet.LineDashed:
		return nsheet.LineDashed
	case sheet.LineDotted:
		return nsheet.LineDotted
	case sheet.LineDouble:
		return nsheet.LineDouble
	}
	return nsheet.LineNone
}
