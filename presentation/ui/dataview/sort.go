// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataview

import (
	"iter"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui/pager"
	"golang.org/x/text/language"
)

var (
	StrSortBy           = i18n.MustVarString("nago.dataview.sort_by", i18n.Values{language.English: "Sort by {name}", language.German: "Nach {name} sortieren"})
	StrSortedAscending  = i18n.MustVarString("nago.dataview.sorted_ascending", i18n.Values{language.English: "Sorted by {name} ascending", language.German: "Aufsteigend nach {name} sortiert"})
	StrSortedDescending = i18n.MustVarString("nago.dataview.sorted_descending", i18n.Values{language.English: "Sorted by {name} descending", language.German: "Absteigend nach {name} sortiert"})
)

// sortState holds the field to sort by and the direction, which all styles share.
type sortState struct {
	field   *core.State[FieldID]
	reverse *core.State[bool]
}

func (t TDataView[E, ID]) sortState(wnd core.Window) sortState {
	return sortState{
		field:   core.StateOf[FieldID](wnd, t.modelOptions.StatePrefix+"-sortByField"),
		reverse: core.StateOf[bool](wnd, t.modelOptions.StatePrefix+"-sortReverse"),
	}
}

// findAll returns the identifiers of all entries, sorted by the chosen field.
func (t TDataView[E, ID]) findAll(s sortState) iter.Seq2[ID, error] {
	if s.field.Get() == "" {
		return t.data.FindAll
	}

	for _, field := range t.data.Fields {
		if field.ID == s.field.Get() && field.Comparator != nil {
			return pager.Sort(t.data.FindByID, t.data.FindAll, field.compare(s.reverse.Get()), pager.SortOptions[E]{})
		}
	}

	return t.data.FindAll
}

// toggle sorts ascending by the given field, or reverses the direction if it is already sorted by it.
func (s sortState) toggle(id FieldID) {
	if s.field.Get() != id {
		s.field.Set(id)
		s.reverse.Set(false)
		return
	}

	s.reverse.Set(!s.reverse.Get())
}

func (s sortState) icon(id FieldID) core.SVG {
	switch {
	case s.field.Get() != id:
		return icons.ArrowUpDown
	case s.reverse.Get():
		return icons.ArrowDown
	default:
		return icons.ArrowUp
	}
}

// label describes the sort control of the given field for screen readers.
func (s sortState) label(wnd core.Window, name string, id FieldID) string {
	switch {
	case s.field.Get() != id:
		return StrSortBy.Get(wnd, i18n.String("name", name))
	case s.reverse.Get():
		return StrSortedDescending.Get(wnd, i18n.String("name", name))
	default:
		return StrSortedAscending.Get(wnd, i18n.String("name", name))
	}
}
