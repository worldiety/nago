package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"iter"
)

type TOptions[Entity data.Aggregate[ID], ID data.IDType] struct {
	title     string
	actions   []core.View // global components to show for the entire crud set, e.g. for custom create action
	create    func(Entity) error
	initEmpty func() (Entity, error)
	findAll   iter.Seq2[Entity, error]
	query     string
	delete    func(Entity) error // this is already an aggregate action
	//aggregateActions []AggregateAction[E]
	//binding          *Binding[E]
	wnd        core.Window
	bnd        *Binding[Entity]
	queryState *core.State[string]
}

func Options[Entity data.Aggregate[ID], ID data.IDType](wnd core.Window, bnd *Binding[Entity]) TOptions[Entity, ID] {
	return TOptions[Entity, ID]{
		wnd: wnd,
		bnd: bnd,
	}
}

func (o TOptions[Entity, ID]) FindAll(it iter.Seq2[Entity, error]) TOptions[Entity, ID] {
	o.findAll = it
	return o
}

func (o TOptions[Entity, ID]) QuickFilterQuery(query string) TOptions[Entity, ID] {
	o.query = query
	return o
}

func (o TOptions[Entity, ID]) datasource() dataSource[Entity, ID] {
	return dataSource[Entity, ID]{
		it:          o.findAll,
		binding:     o.bnd,
		sortByField: nil,
		query:       o.queryState.Get(),
	}
}

func (o TOptions[Entity, ID]) Title(title string) TOptions[Entity, ID] {
	o.title = title
	return o
}
