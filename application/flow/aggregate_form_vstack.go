// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import "go.wdy.de/nago/presentation/ui"

var (
	_ FormViewGroup  = (*FormVStack)(nil)
	_ Alignable      = (*FormVStack)(nil)
	_ Borderable     = (*FormVStack)(nil)
	_ Frameable      = (*FormVStack)(nil)
	_ Backgroundable = (*FormVStack)(nil)
	_ Paddable       = (*FormVStack)(nil)
)

type FormVStack struct {
	*baseViewGroup
	alignment       ui.Alignment
	gap             ui.Length
	frame           ui.Frame
	border          ui.Border
	backgroundColor ui.Color
	padding         ui.Padding
}

func NewFormVStack(id ViewID) *FormVStack {
	return &FormVStack{baseViewGroup: &baseViewGroup{id: id}}
}

func (f *FormVStack) Clone() FormView {
	return &FormVStack{
		baseViewGroup:   f.baseViewGroup.clone(),
		alignment:       f.alignment,
		gap:             f.gap,
		frame:           f.frame,
		border:          f.border,
		backgroundColor: f.backgroundColor,
		padding:         f.padding,
	}
}

func (f *FormVStack) Alignment() ui.Alignment {
	return f.alignment
}

func (f *FormVStack) SetAlignment(a ui.Alignment) {
	f.alignment = a
}

func (f *FormVStack) Gap() ui.Length {
	return f.gap
}

func (f *FormVStack) SetGap(l ui.Length) {
	f.gap = l
}

func (f *FormVStack) Frame() ui.Frame {
	return f.frame
}

func (f *FormVStack) SetFrame(frame ui.Frame) {
	f.frame = frame
}

func (f *FormVStack) Border() ui.Border {
	return f.border
}

func (f *FormVStack) SetBorder(border ui.Border) {
	f.border = border
}

func (f *FormVStack) BackgroundColor() ui.Color {
	return f.backgroundColor
}

func (f *FormVStack) SetBackgroundColor(color ui.Color) {
	f.backgroundColor = color
}

func (f *FormVStack) Padding() ui.Padding {
	return f.padding
}

func (f *FormVStack) SetPadding(padding ui.Padding) {
	f.padding = padding
}
