// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

type FormCheckbox struct {
	id             ViewID
	structType     TypeID
	field          FieldID
	label          string
	supportingText string
}

func NewFormCheckbox(id ViewID, structType TypeID, field FieldID) *FormCheckbox {
	return &FormCheckbox{
		id:         id,
		structType: structType,
		field:      field,
	}
}

func (f *FormCheckbox) Identity() ViewID {
	return f.id
}

func (f *FormCheckbox) Clone() FormView {
	c := *f
	return &c
}

func (f *FormCheckbox) Label() string {
	return f.label
}

func (f *FormCheckbox) Field() FieldID {
	return f.field
}

func (f *FormCheckbox) SetLabel(s string) {
	f.label = s
}

func (f *FormCheckbox) SupportingText() string {
	return f.supportingText
}

func (f *FormCheckbox) SetSupportingText(s string) {
	f.supportingText = s
}
