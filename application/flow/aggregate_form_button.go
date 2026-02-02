// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"slices"

	"go.wdy.de/nago/presentation/ui"
)

var (
	_ Frameable = (*FormButton)(nil)
)

type FormButton struct {
	id          ViewID
	title       string
	visibleExpr Expression // if empty, its visible
	action      []Expression
	enabled     Expression // if empty, its enabled
	style       ui.ButtonStyle
	frame       ui.Frame
}

func NewFormButton(id ViewID, title string, style ui.ButtonStyle) *FormButton {
	return &FormButton{
		id:    id,
		title: title,
		style: style,
	}
}

func (f *FormButton) SetVisibleExpr(expr Expression) {
	f.visibleExpr = expr
}

func (f *FormButton) VisibleExpr() Expression {
	return f.visibleExpr
}

func (f *FormButton) Clone() FormView {
	c := *f
	c.action = slices.Clone(f.action)
	return &c
}

func (f *FormButton) Title() string {
	return f.title
}

func (f *FormButton) SetActionExpr(exprs ...Expression) {
	f.action = exprs
}

func (f *FormButton) ActionExpr() []Expression {
	return f.action
}

func (f *FormButton) SetEnabledExpr(expr Expression) {
	f.enabled = expr
}

func (f *FormButton) EnabledExpr() Expression {
	return f.enabled
}

func (f *FormButton) Style() ui.ButtonStyle {
	return f.style
}

func (f *FormButton) SetStyle(style ui.ButtonStyle) {
	f.style = style
}

func (f *FormButton) Identity() ViewID {
	return f.id
}

func (f *FormButton) Frame() ui.Frame {
	return f.frame
}

func (f *FormButton) SetFrame(frame ui.Frame) {
	f.frame = frame
}
