// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import "go.wdy.de/nago/presentation/ui"

var (
	_ FormViewGroup = (*FormHStack)(nil)
	_ Alignable     = (*FormHStack)(nil)
)

type FormHStack struct {
	*baseViewGroup
	alignment ui.Alignment
	gap       ui.Length
}

func NewFormHStack(id ViewID) *FormHStack {
	return &FormHStack{baseViewGroup: &baseViewGroup{id: id}}
}

func (f *FormHStack) Clone() FormView {
	return &FormHStack{
		baseViewGroup: f.baseViewGroup.clone(),
		alignment:     f.alignment,
		gap:           f.gap,
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
