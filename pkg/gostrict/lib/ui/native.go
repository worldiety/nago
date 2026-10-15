// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	"context"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/lib/ui/sheet"
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
	nsheet "go.wdy.de/nago/presentation/ui/sheet"
)

// The handles of the package hold the values of nago, e.g. a TText holds a ui.TText of nago. The functions convert
// the values of the package into the values of nago and back.

func wrap(v any) js.Ref {
	return js.NewRef(v)
}

// get returns the value of nago, which the handle holds, or its zero value, e.g. for the zero handle, which a failed
// callback returns.
func get[T any](ref js.Ref) T {
	v, _ := ref.Value().(T)
	return v
}

// viewOf returns the view of nago of a view or nil.
func viewOf(v View) core.View {
	h, ok := v.(interface{ HostRef() js.Ref })
	if !ok {
		return nil
	}
	cv, _ := h.HostRef().Value().(core.View)
	return cv
}

func viewsOf(views []View) []core.View {
	res := make([]core.View, 0, len(views))
	for _, v := range views {
		res = append(res, viewOf(v))
	}
	return res
}

// Render calls the exported function of the instance, which takes the window and returns its view, e.g.
// func Render(wnd ui.Window) ui.View, and returns the view of nago. A failure is shown as a text.
func Render(ctx context.Context, in gostrict.Instance, export string, wnd core.Window) core.View {
	res, err := in.Call(ctx, export, wrap(wnd))
	if err != nil {
		return nui.Text(err.Error())
	}
	var v View
	if err := gostrict.FromNeutral(res[0], &v); err != nil {
		return nui.Text(err.Error())
	}
	return viewOf(v)
}

// WindowOf returns the handle of a window of nago, e.g. for a program, which runs natively.
func WindowOf(wnd core.Window) Window {
	return Window{ref: wrap(wnd)}
}

// Core returns the view of nago of a view, e.g. of a program, which runs natively.
func Core(v View) core.View {
	return viewOf(v)
}

func (w Window) nago() core.Window {
	return get[core.Window](w.ref)
}

func (s StringState) nago() *core.State[string] {
	return get[*core.State[string]](s.ref)
}

func (s BoolState) nago() *core.State[bool] {
	return get[*core.State[bool]](s.ref)
}

func (p Padding) nago() nui.Padding {
	return nui.Padding{Top: nui.Length(p.Top), Left: nui.Length(p.Left), Right: nui.Length(p.Right), Bottom: nui.Length(p.Bottom)}
}

func paddingFrom(p nui.Padding) Padding {
	return Padding{Top: Length(p.Top), Left: Length(p.Left), Right: Length(p.Right), Bottom: Length(p.Bottom)}
}

func (f Frame) nago() nui.Frame {
	return nui.Frame{MinWidth: nui.Length(f.MinWidth), MaxWidth: nui.Length(f.MaxWidth), MinHeight: nui.Length(f.MinHeight), MaxHeight: nui.Length(f.MaxHeight), Width: nui.Length(f.Width), Height: nui.Length(f.Height), FlexGrow: f.FlexGrow, FlexPreventShrink: f.FlexPreventShrink}
}

func frameFrom(f nui.Frame) Frame {
	return Frame{MinWidth: Length(f.MinWidth), MaxWidth: Length(f.MaxWidth), MinHeight: Length(f.MinHeight), MaxHeight: Length(f.MaxHeight), Width: Length(f.Width), Height: Length(f.Height), FlexGrow: f.FlexGrow, FlexPreventShrink: f.FlexPreventShrink}
}

func (f Font) nago() nui.Font {
	return nui.Font{Name: nui.FontName(f.Name), Size: nui.Length(f.Size), Style: nui.FontStyle(f.Style), Weight: nui.FontWeight(f.Weight), LineHeight: nui.LineHeight(f.LineHeight)}
}

func fontFrom(f nui.Font) Font {
	return Font{Name: FontName(f.Name), Size: Length(f.Size), Style: FontStyle(f.Style), Weight: FontWeight(f.Weight), LineHeight: LineHeight(f.LineHeight)}
}

func (s Shadow) nago() nui.Shadow {
	return nui.Shadow{Color: nui.Color(s.Color), Radius: nui.Length(s.Radius), X: nui.Length(s.X), Y: nui.Length(s.Y)}
}

func shadowFrom(s nui.Shadow) Shadow {
	return Shadow{Color: Color(s.Color), Radius: Length(s.Radius), X: Length(s.X), Y: Length(s.Y)}
}

func (b Border) nago() nui.Border {
	return nui.Border{
		TopLeftRadius: nui.Length(b.TopLeftRadius), TopRightRadius: nui.Length(b.TopRightRadius),
		BottomLeftRadius: nui.Length(b.BottomLeftRadius), BottomRightRadius: nui.Length(b.BottomRightRadius),
		LeftWidth: nui.Length(b.LeftWidth), TopWidth: nui.Length(b.TopWidth), RightWidth: nui.Length(b.RightWidth), BottomWidth: nui.Length(b.BottomWidth),
		LeftColor: nui.Color(b.LeftColor), TopColor: nui.Color(b.TopColor), RightColor: nui.Color(b.RightColor), BottomColor: nui.Color(b.BottomColor),
		BoxShadow: b.BoxShadow.nago(), BorderStyle: nui.BorderStyle(b.BorderStyle),
	}
}

func borderFrom(b nui.Border) Border {
	return Border{
		TopLeftRadius: Length(b.TopLeftRadius), TopRightRadius: Length(b.TopRightRadius),
		BottomLeftRadius: Length(b.BottomLeftRadius), BottomRightRadius: Length(b.BottomRightRadius),
		LeftWidth: Length(b.LeftWidth), TopWidth: Length(b.TopWidth), RightWidth: Length(b.RightWidth), BottomWidth: Length(b.BottomWidth),
		LeftColor: Color(b.LeftColor), TopColor: Color(b.TopColor), RightColor: Color(b.RightColor), BottomColor: Color(b.BottomColor),
		BoxShadow: shadowFrom(b.BoxShadow), BorderStyle: BorderStyle(b.BorderStyle),
	}
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
		BackgroundColor: nui.Color(s.BackgroundColor),
		Color:           nui.Color(s.Color),
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
	return nsheet.Line{Color: nui.Color(l.Color), Width: l.Width, Style: lineStyleOf(l.Style)}
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
