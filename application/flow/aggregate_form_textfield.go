// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import "go.wdy.de/nago/presentation/ui"

var (
	_ FormView  = (*FormTextField)(nil)
	_ Frameable = (*FormTextField)(nil)
)

type FormTextField struct {
	id             ViewID
	structType     TypeID
	field          FieldID
	frame          ui.Frame
	visibleExpr    Expression
	lines          int
	label          string
	supportingText string
}

func NewFormTextField(id ViewID, structType TypeID, field FieldID) *FormTextField {
	return &FormTextField{
		id:         id,
		structType: structType,
		field:      field,
	}
}

func (f *FormTextField) Identity() ViewID {
	return f.id
}

func (f *FormTextField) VisibleExpr() Expression {
	return f.visibleExpr
}

func (f *FormTextField) SetVisibleExpr(expression Expression) {
	f.visibleExpr = expression
}

func (f *FormTextField) Clone() FormView {
	c := *f
	return &c
}

func (f *FormTextField) Frame() ui.Frame {
	return f.frame
}

func (f *FormTextField) SetFrame(frame ui.Frame) {
	f.frame = frame
}

func (f *FormTextField) Label() string {
	return f.label
}

func (f *FormTextField) SetLabel(s string) {
	f.label = s
}

func (f *FormTextField) SupportingText() string {
	return f.supportingText
}

func (f *FormTextField) SetSupportingText(s string) {
	f.supportingText = s
}

func (f *FormTextField) Lines() int {
	return f.lines
}

func (f *FormTextField) SetLines(lines int) {
	f.lines = lines
}

func (f *FormTextField) StructType() TypeID {
	return f.structType
}

func (f *FormTextField) Field() FieldID {
	return f.field
}
