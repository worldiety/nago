// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import "go.wdy.de/nago/presentation/ui"

var (
	_ FormView  = (*FormHLine)(nil)
	_ Paddable  = (*FormHLine)(nil)
	_ Frameable = (*FormHLine)(nil)
)

type FormHLine struct {
	expr    Expression
	id      ViewID
	frame   ui.Frame
	padding ui.Padding
}

func NewFormHLine(id ViewID) *FormHLine {
	return &FormHLine{id: id}
}

func (f *FormHLine) Identity() ViewID {
	return f.id
}

func (f *FormHLine) VisibleExpr() Expression {
	return f.expr
}

func (f *FormHLine) SetVisibleExpr(expression Expression) {
	f.expr = expression
}

func (f *FormHLine) Clone() FormView {
	c := *f
	return &c
}

func (f *FormHLine) Frame() ui.Frame {
	return f.frame
}

func (f *FormHLine) SetFrame(frame ui.Frame) {
	f.frame = frame
}

func (f *FormHLine) Padding() ui.Padding {
	return f.padding
}

func (f *FormHLine) SetPadding(padding ui.Padding) {
	f.padding = padding
}
