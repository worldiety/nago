// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"iter"
	"sync/atomic"

	"go.wdy.de/nago/pkg/xslices"
)

type Form struct {
	parent      *Workspace
	id          FormID
	name        atomic.Pointer[string]
	description atomic.Pointer[string]
	elements    atomic.Pointer[xslices.Slice[FormElement]]
}

func (f *Form) Elements() iter.Seq[FormElement] {
	pSlice := f.elements.Load()
	if pSlice == nil {
		return xslices.Slice[FormElement]{}.All()
	}

	return pSlice.All()
}

func (f *Form) Identity() FormID {
	return f.id
}

func (f *Form) Name() string {
	return *f.name.Load()
}

func (f *Form) Description() string {
	return *f.description.Load()
}

func (f *Form) formElement() {}

type FormElement interface {
	Identity() ElementID
	formElement()
	Label() string
	SupportingText() string
}

type ElementID string

type FormCard struct {
	parent         *Form
	id             ElementID
	label          atomic.Pointer[string]
	supportingText atomic.Pointer[string]
	elements       atomic.Pointer[xslices.Slice[FormElement]]
}

func (f *FormCard) formElement() {}

func (f *FormCard) Elements() iter.Seq[FormElement] {
	return f.elements.Load().All()
}

func (f *FormCard) Label() string {
	return *f.label.Load()
}

func (f *FormCard) SupportingText() string {
	return *f.supportingText.Load()
}

func (f *FormCard) Identity() ElementID {
	return f.id
}

type FormCheckbox struct {
	id             ElementID
	field          FieldID
	label          atomic.Pointer[string]
	supportingText atomic.Pointer[string]
	visible        VisibilityValueRule
}

type VisibilityValueRule struct {
	Field      FieldID
	Value      string
	Comparator Comparator
}

type Comparator int

const (
	CmpEqual Comparator = iota + 1
	CmpNotEqual
	CmpLargerThan
	CmpSmallerThan
	CmpRegExp
)
