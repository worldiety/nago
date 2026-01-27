// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"iter"
	"maps"
	"slices"
	"strings"

	"go.wdy.de/nago/pkg/xmaps"
)

type Forms struct {
	forms map[FormID]*Form
}

func NewForms() *Forms {
	return &Forms{forms: make(map[FormID]*Form)}
}

func (forms *Forms) ByID(id FormID) (*Form, bool) {
	form, ok := forms.forms[id]
	return form, ok
}

func (forms *Forms) Remove(id FormID) (*Form, bool) {
	form, ok := forms.forms[id]
	if ok {
		delete(forms.forms, id)
	}

	return form, ok
}

func (forms *Forms) ByView(id ViewID) (*Form, bool) {
	for _, form := range forms.forms {
		if _, ok := FindElementByID(form.Root, id); ok {
			return form, true
		}
	}

	return nil, false
}

func (forms *Forms) ViewByID(id ViewID) (FormView, bool) {
	for _, form := range forms.forms {
		if v, ok := FindElementByID(form.Root, id); ok {
			return v, true
		}
	}

	return nil, false
}

func (forms *Forms) AddForm(form *Form) {
	forms.forms[form.Identity()] = form
}

func (forms *Forms) Len() int {
	return len(forms.forms)
}

func (forms *Forms) All() iter.Seq[*Form] {
	return slices.Values(slices.SortedFunc(maps.Values(forms.forms), func(a, b *Form) int { return strings.Compare(string(a.Name()), string(b.Name())) }))
}

func (forms *Forms) Clone() *Forms {
	return &Forms{forms: xmaps.Clone(forms.forms)}
}
