// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import "go.wdy.de/nago/presentation/ui"

var (
	_ FormViewGroup = (*FormVStack)(nil)
	_ Alignable     = (*FormVStack)(nil)
)

type FormVStack struct {
	*baseViewGroup
	alignment ui.Alignment
}

func NewFormVStack(id ViewID) *FormVStack {
	return &FormVStack{baseViewGroup: &baseViewGroup{id: id}}
}

func (f *FormVStack) Clone() FormView {
	return &FormVStack{baseViewGroup: f.baseViewGroup.clone()}
}

func (f *FormVStack) Alignment() ui.Alignment {
	return f.alignment
}

func (f *FormVStack) SetAlignment(a ui.Alignment) {
	f.alignment = a
}
