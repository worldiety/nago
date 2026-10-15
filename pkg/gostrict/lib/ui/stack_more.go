// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import nui "go.wdy.de/nago/presentation/ui"

// StackLayout is the layout of the children of a stack, see TStack.Layout.
type StackLayout int32

const (
	// StackLayoutAuto lays out the children in a column on a small window and in a row otherwise.
	StackLayoutAuto StackLayout = 0
	// StackLayoutVertical lays out the children in a column.
	StackLayoutVertical StackLayout = 1
	// StackLayoutHorizontal lays out the children in a row.
	StackLayoutHorizontal StackLayout = 2
	// StackLayoutWrap lays out the children in rows, which wrap into further rows, if they do not fit.
	StackLayoutWrap StackLayout = 3
	// StackLayoutFlex is the same as StackLayoutWrap.
	StackLayoutFlex StackLayout = 3
)

// StylePreset applies a predefined style preset, e.g. ButtonStylePrimary, to the stack, which then looks like a button.
//
//gostrict:host
func (c TStack) StylePreset(preset ButtonStyle) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).StylePreset(nui.ButtonStyle(preset)))}
}

// Position sets the position of the stack within its parent layout.
//
//gostrict:host
func (c TStack) Position(position Position) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Position(position.nago()))}
}

// Animation sets the animation of the stack.
//
//gostrict:host
func (c TStack) Animation(animation Animation) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Animation(nui.Animation(animation)))}
}

// Transformation moves, scales and rotates the stack.
//
//gostrict:host
func (c TStack) Transformation(transformation Transformation) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Transformation(transformation.nago()))}
}

// Outline sets the outline of the stack.
//
//gostrict:host
func (c TStack) Outline(outline Outline) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Outline(outline.nago()))}
}

// HoveredOutline sets the outline, while the pointer is over the stack.
//
//gostrict:host
func (c TStack) HoveredOutline(outline Outline) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).HoveredOutline(outline.nago()))}
}

// PressedOutline sets the outline, while the stack is pressed.
//
//gostrict:host
func (c TStack) PressedOutline(outline Outline) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).PressedOutline(outline.nago()))}
}

// FocusedOutline sets the outline, while the stack has the focus.
//
//gostrict:host
func (c TStack) FocusedOutline(outline Outline) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).FocusedOutline(outline.nago()))}
}

// Layout sets the layout of the children, e.g. StackLayoutHorizontal.
//
//gostrict:host
func (c TStack) Layout(layout StackLayout) TStack {
	return TStack{ref: wrap(get[nui.TStack](c.ref).Layout(nui.StackLayout(layout)))}
}
