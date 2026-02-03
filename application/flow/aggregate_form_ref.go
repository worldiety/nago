// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

type FormRef struct {
	id          ViewID
	ref         FormID
	visibleExpr Expression
}

func NewFormRef(id ViewID, ref FormID) *FormRef {
	return &FormRef{
		id:  id,
		ref: ref,
	}
}

func (f *FormRef) Ref() FormID {
	return f.ref
}

func (f *FormRef) SetVisibleExpr(expr Expression) {
	f.visibleExpr = expr
}

func (f *FormRef) VisibleExpr() Expression {
	return f.visibleExpr
}

func (f *FormRef) Clone() FormView {
	c := *f
	return &c
}

func (f *FormRef) Identity() ViewID {
	return f.id
}
