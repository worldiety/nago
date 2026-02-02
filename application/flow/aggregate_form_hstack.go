// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import "go.wdy.de/nago/presentation/ui"

var (
	_ FormViewGroup  = (*FormHStack)(nil)
	_ Alignable      = (*FormHStack)(nil)
	_ Borderable     = (*FormHStack)(nil)
	_ Frameable      = (*FormHStack)(nil)
	_ Backgroundable = (*FormHStack)(nil)
	_ Paddable       = (*FormHStack)(nil)
)

type FormHStack struct {
	*baseViewGroup
	alignment       ui.Alignment
	gap             ui.Length
	frame           ui.Frame
	border          ui.Border
	backgroundColor ui.Color
	padding         ui.Padding
}

func NewFormHStack(id ViewID) *FormHStack {
	return &FormHStack{baseViewGroup: &baseViewGroup{id: id}}
}

func (f *FormHStack) Clone() FormView {
	return &FormHStack{
		baseViewGroup:   f.baseViewGroup.clone(),
		alignment:       f.alignment,
		gap:             f.gap,
		frame:           f.frame,
		border:          f.border,
		backgroundColor: f.backgroundColor,
		padding:         f.padding,
	}
}

func (f *FormHStack) Alignment() ui.Alignment {
	return f.alignment
}

func (f *FormHStack) SetAlignment(a ui.Alignment) {
	f.alignment = a
}

func (f *FormHStack) Gap() ui.Length {
	return f.gap
}

func (f *FormHStack) SetGap(l ui.Length) {
	f.gap = l
}

func (f *FormHStack) Frame() ui.Frame {
	return f.frame
}

func (f *FormHStack) SetFrame(frame ui.Frame) {
	f.frame = frame
}

func (f *FormHStack) Border() ui.Border {
	return f.border
}

func (f *FormHStack) SetBorder(border ui.Border) {
	f.border = border
}

func (f *FormHStack) BackgroundColor() ui.Color {
	return f.backgroundColor
}

func (f *FormHStack) SetBackgroundColor(color ui.Color) {
	f.backgroundColor = color
}

func (f *FormHStack) Padding() ui.Padding {
	return f.padding
}

func (f *FormHStack) SetPadding(padding ui.Padding) {
	f.padding = padding
}
