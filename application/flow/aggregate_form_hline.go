// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

var (
	_ FormView = (*FormHLine)(nil)
)

type FormHLine struct {
	expr Expression
	id   ViewID
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
