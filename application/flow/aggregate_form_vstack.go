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
	gap       ui.Length
}

func NewFormVStack(id ViewID) *FormVStack {
	return &FormVStack{baseViewGroup: &baseViewGroup{id: id}}
}

func (f *FormVStack) Clone() FormView {
	return &FormVStack{
		baseViewGroup: f.baseViewGroup.clone(),
		alignment:     f.alignment,
		gap:           f.gap,
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
