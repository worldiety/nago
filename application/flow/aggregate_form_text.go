// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

type FormText struct {
	id          ViewID
	value       string
	style       FormTextStyle
	visibleExpr Expression
}

func NewFormText(id ViewID, value string, style FormTextStyle) *FormText {
	return &FormText{
		id:    id,
		value: value,
		style: style,
	}
}

func (f *FormText) SetVisibleExpr(expr Expression) {
	f.visibleExpr = expr
}

func (f *FormText) VisibleExpr() Expression {
	return f.visibleExpr
}

func (f *FormText) Clone() FormView {
	c := *f
	return &c
}

func (f *FormText) Value() string {
	return f.value
}

func (f *FormText) Style() FormTextStyle {
	return f.style
}

func (f *FormText) Identity() ViewID {
	return f.id
}
