package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"iter"
)

type sortDir bool

const (
	asc  sortDir = true
	desc sortDir = false
)

type TOptions[Entity data.Aggregate[ID], ID data.IDType] struct {
	title            string
	actions          []core.View // global components to show for the entire crud set, e.g. for custom create action
	findAll          iter.Seq2[Entity, error]
	wnd              core.Window
	bnd              *Binding[Entity]
	queryState       *core.State[string]
	sortByFieldState *core.State[*Field[Entity]]
	sortDirState     *core.State[sortDir]
}

func Options[Entity data.Aggregate[ID], ID data.IDType](bnd *Binding[Entity]) TOptions[Entity, ID] {
	wnd := bnd.wnd
	return TOptions[Entity, ID]{
		wnd:              wnd,
		bnd:              bnd,
		queryState:       core.AutoState[string](wnd),
		sortByFieldState: core.AutoState[*Field[Entity]](wnd),
		sortDirState:     core.AutoState[sortDir](wnd),
	}
}

func (o TOptions[Entity, ID]) FindAll(it iter.Seq2[Entity, error]) TOptions[Entity, ID] {
	o.findAll = it
	return o
}

func (o TOptions[Entity, ID]) Actions(actions ...core.View) TOptions[Entity, ID] {
	o.actions = actions
	return o
}

func (o TOptions[Entity, ID]) datasource() dataSource[Entity, ID] {
	return dataSource[Entity, ID]{
		it:          o.findAll,
		binding:     o.bnd,
		sortByField: o.sortByFieldState.Get(),
		sortOrder:   o.sortDirState.Get(),
		query:       o.queryState.Get(),
	}
}

func (o TOptions[Entity, ID]) Title(title string) TOptions[Entity, ID] {
	o.title = title
	return o
}
