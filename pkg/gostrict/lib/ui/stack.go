// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
)

// TStack lays out its children in a row or a column, see VStack and HStack.
type TStack struct{ ref js.Ref }

func (TStack) isView() {}

// Stack returns a stack, whose layout the children decide.
//
//gostrict:host
func Stack(children ...View) TStack {
	return TStack{ref: wrap(nui.Stack(viewsOf(children)...))}
}

// HStack returns a row of the children.
//
//gostrict:host
func HStack(children ...View) TStack {
	return TStack{ref: wrap(nui.HStack(viewsOf(children)...))}
}

// VStack returns a column of the children.
//
//gostrict:host
func VStack(children ...View) TStack {
	return TStack{ref: wrap(nui.VStack(viewsOf(children)...))}
}

// Responsive changes the stack with fn, which gets the window, e.g. to lay it out by the size of the window.
//
//gostrict:host
func (c TStack) Responsive(fn func(wnd Window, stack TStack) TStack) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Responsive(func(wnd core.Window, s nui.TStack) nui.TStack {
		return get[nui.TStack](fn(Window{ref: wrap(wnd)}, TStack{ref: wrap(s)}).ref)
	}))}
}

// With changes the stack with fn.
//
//gostrict:host
func (c TStack) With(fn func(stack TStack) TStack) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).With(func(s nui.TStack) nui.TStack {
		return get[nui.TStack](fn(TStack{ref: wrap(s)}).ref)
	}))}
}

// NoClip lets the children overflow the stack.
//
//gostrict:host
func (c TStack) NoClip(b bool) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).NoClip(b))}
}

// Font sets the font of the texts.
//
//gostrict:host
func (c TStack) Font(font Font) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Font(font.nago()))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TStack) AccessibilityLabel(label string) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).AccessibilityLabel(label).(nui.TStack))}
}

// Visible shows or hides the stack.
//
//gostrict:host
func (c TStack) Visible(visible bool) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Visible(visible).(nui.TStack))}
}

// Action calls f, when the user clicks the stack.
//
//gostrict:host
func (c TStack) Action(f func()) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Action(f))}
}

// Key sets a key, which triggers the action.
//
//gostrict:host
func (c TStack) Key(name, id string) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Key(name, id))}
}

// HRef makes the stack a link.
//
//gostrict:host
func (c TStack) HRef(url URI) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).HRef(core.URI(url)))}
}

// Target sets the target of the link.
//
//gostrict:host
func (c TStack) Target(target string) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Target(target))}
}

// Wrap wraps the children into further rows or columns.
//
//gostrict:host
func (c TStack) Wrap(b bool) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Wrap(b))}
}

// Enabled enables or disables the action.
//
//gostrict:host
func (c TStack) Enabled(enabled bool) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Enabled(enabled))}
}

// ID sets the id.
//
//gostrict:host
func (c TStack) ID(id string) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).ID(id))}
}

// Opacity sets the opacity from 0 to 1.
//
//gostrict:host
func (c TStack) Opacity(opacity float64) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Opacity(opacity))}
}

// TextColor sets the color of the texts.
//
//gostrict:host
func (c TStack) TextColor(textColor Color) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).TextColor(nui.Color(textColor)))}
}

// Append appends the children.
//
//gostrict:host
func (c TStack) Append(children ...View) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Append(viewsOf(children)...))}
}

// BackgroundColor sets the color of the background.
//
//gostrict:host
func (c TStack) BackgroundColor(color Color) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).BackgroundColor(nui.Color(color)))}
}

// HoveredBackgroundColor sets the color of the background, while the pointer is over the stack.
//
//gostrict:host
func (c TStack) HoveredBackgroundColor(backgroundColor Color) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).HoveredBackgroundColor(nui.Color(backgroundColor)))}
}

// PressedBackgroundColor sets the color of the background, while the stack is pressed.
//
//gostrict:host
func (c TStack) PressedBackgroundColor(backgroundColor Color) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).PressedBackgroundColor(nui.Color(backgroundColor)))}
}

// FocusedBackgroundColor sets the color of the background, while the stack has the focus.
//
//gostrict:host
func (c TStack) FocusedBackgroundColor(backgroundColor Color) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).FocusedBackgroundColor(nui.Color(backgroundColor)))}
}

// Alignment aligns the children.
//
//gostrict:host
func (c TStack) Alignment(alignment Alignment) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Alignment(nui.Alignment(alignment)))}
}

// Frame sets the size.
//
//gostrict:host
func (c TStack) Frame(frame Frame) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Frame(frame.nago()).(nui.TStack))}
}

// WithFrame changes the size with fn, which gets the current one.
//
//gostrict:host
func (c TStack) WithFrame(fn func(Frame) Frame) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(nui.TStack))}
}

// Gap sets the space between the children.
//
//gostrict:host
func (c TStack) Gap(gap Length) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Gap(nui.Length(gap)))}
}

// Padding sets the padding.
//
//gostrict:host
func (c TStack) Padding(padding Padding) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Padding(padding.nago()).(nui.TStack))}
}

// WithPadding sets the padding.
//
//gostrict:host
func (c TStack) WithPadding(padding Padding) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).WithPadding(padding.nago()))}
}

// Border sets the border.
//
//gostrict:host
func (c TStack) Border(border Border) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Border(border.nago()).(nui.TStack))}
}

// HoveredBorder sets the border, while the pointer is over the stack.
//
//gostrict:host
func (c TStack) HoveredBorder(border Border) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).HoveredBorder(border.nago()))}
}

// PressedBorder sets the border, while the stack is pressed.
//
//gostrict:host
func (c TStack) PressedBorder(border Border) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).PressedBorder(border.nago()))}
}

// FocusedBorder sets the border, while the stack has the focus.
//
//gostrict:host
func (c TStack) FocusedBorder(border Border) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).FocusedBorder(border.nago()))}
}

// HtmlTag sets the tag of HTML, e.g. main.
//
//gostrict:host
func (c TStack) HtmlTag(tag string) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).HtmlTag(tag))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TStack) FullWidth() TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).FullWidth())}
}
