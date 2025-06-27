// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package pager

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"iter"
	"slices"
)

type ModelOptions struct {
	StatePrefix string
	PageSize    int // Defaults to whatever [data.PaginateOptions] defines.
}

// Model is a simplified aggregate of combining paging, filtering and selection all at once.
type Model[E data.Aggregate[ID], ID ~string] struct {
	Window core.Window
	// Query represents the build-in rquery.SimplePredicate filter input to create a visible subset.
	Query *core.State[string]

	// PageIdx is the active 0-based page offset. Use this state as the page index state for [Pager].
	PageIdx *core.State[int]

	// Selections contain the relation between each identity and if it has been actually selected.
	// Use each state to connect it to a [ui.Checkbox].
	Selections map[ID]*core.State[bool]

	// SelectionCount is the number of selection states whose value is true, independent of the current subset.
	// Use this to communicate the total state of selection, even if parts are hidden.
	SelectionCount int

	// SelectSubset is a two-way state flag, which indicates if the current subset is entirely selected or not.
	// Vice versa, if triggered and notified, it will update the current selection accordingly.
	SelectSubset *core.State[bool]

	// Page contains the actually loaded data set of unmarshalled entities which shall be displayed.
	Page data.Page[E]

	// UnselectAll removes the entire selection, independent of any active subset.
	UnselectAll func()
}

// NewModel creates a new model which provides a bunch of reasonable defaults, like quick filter, paging and selection.
// Inspect the field documentation of [Model] to see what it provides and how it helps you to avoid standard
// legwork. It also tries its best to avoid as much as memory consumption as possible; however, we still have
// various limitations:
//   - if a query is provided, a filter is applied which requires an O(N) unmarshalling run over all entities. But to
//     avoid expensive memory usage, each entry is discarded after read, thus resulting in O(1) temporary memory usage.
//   - all identifiers from the iterator are stored in memory. What is even worse, a state for each is created and
//     held in memory.
//   - there are internal update routines, which may currently cause some O(Nˆ2) loops. For large datasets, this may
//     break your machine.
func NewModel[E data.Aggregate[ID], ID ~string](wnd core.Window, findByID data.ByIDFinder[E, ID], it iter.Seq2[ID, error], opts ModelOptions) (Model[E, ID], error) {
	var model Model[E, ID]
	allEntityIdents, err := xslices.Collect2(it)
	if err != nil {
		return model, fmt.Errorf("cannot collect all identifiers: %w", err)
	}

	model.Query = core.StateOf[string](wnd, opts.StatePrefix+"-query")
	model.PageIdx = core.StateOf[int](wnd, opts.StatePrefix+"-pageIdx")

	type tableHolder struct {
		idents []ID
	}

	filterOpts := data.FilterOptions[E, ID]{}
	allEntityIdentsInSubset := core.AutoState[*tableHolder](wnd).Init(func() *tableHolder {
		return &tableHolder{}
	})
	allEntityIdentsInSubset.Get().idents = allEntityIdentsInSubset.Get().idents[:0]
	if model.Query.Get() != "" {
		p := rquery.SimplePredicate[E](model.Query.Get())
		filterOpts.Accept = func(u E) bool {
			if p(u) {
				s := allEntityIdentsInSubset.Get()
				s.idents = append(s.idents, u.Identity())
				return true
			}

			return false
		}
	} else {
		allEntityIdentsInSubset.Get().idents = slices.Clone(allEntityIdents)
	}

	page, err := data.FilterAndPaginate[E, ID](
		findByID,
		xslices.Values2[[]ID, ID, error](allEntityIdents),
		filterOpts,
		data.PaginateOptions{
			PageIdx:  model.PageIdx.Get(),
			PageSize: opts.PageSize,
		},
	)

	if err != nil {
		return model, fmt.Errorf("cannot filter and paginate: %w", err)
	}

	model.Page = page

	var recalcSelectedAll func()
	allTableSelected := core.AutoState[bool](wnd).Observe(func(newValue bool) {
		for _, ident := range allEntityIdentsInSubset.Get().idents {
			core.StateOf[bool](wnd, opts.StatePrefix+"-checkbox-"+string(ident)).Set(newValue)
		}
		recalcSelectedAll()
	})

	model.SelectSubset = allTableSelected

	recalcSelectedAll = func() {
		allSelected := true
		for _, id := range allEntityIdentsInSubset.Get().idents {
			if !core.StateOf[bool](wnd, opts.StatePrefix+"-checkbox-"+string(id)).Get() {
				allSelected = false
				break
			}
		}

		if len(allEntityIdentsInSubset.Get().idents) == 0 {
			allSelected = false
		}

		allTableSelected.Set(allSelected)
	}

	recalcSelectedAll()

	countSelection := func() int {
		c := 0
		for _, ident := range allEntityIdents {
			if core.StateOf[bool](wnd, opts.StatePrefix+"-checkbox-"+string(ident)).Get() {
				c++
			}
		}

		return c
	}

	model.SelectionCount = countSelection()

	// always allocate check states for the entire set of users so that we will never loose them, e.g. if not visible
	checkboxStates := map[ID]*core.State[bool]{}
	for _, ident := range allEntityIdents {
		checkboxStates[ident] = core.StateOf[bool](wnd, opts.StatePrefix+"-checkbox-"+string(ident)).Observe(func(newValue bool) {
			recalcSelectedAll()
		})
	}

	model.Selections = checkboxStates
	model.UnselectAll = func() {
		for _, c := range checkboxStates {
			c.Set(false)
		}

		for _, c := range checkboxStates {
			c.Notify()
			break
		}
	}
	return model, nil
}

// PageString returns a formatted localized string like "1-50 von 123 Einträgen".
func (m Model[E, ID]) PageString() string {
	return fmt.Sprintf("%d-%d von %d Einträgen", m.Page.PageIdx*m.Page.PageSize+1, m.Page.PageIdx*m.Page.PageSize+m.Page.PageSize, m.Page.Total)
}
