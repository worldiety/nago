package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/core"
)

type Options[Entity data.Aggregate[ID], ID data.IDType] struct {
	title     string
	actions   []core.View // global components to show for the entire crud set, e.g. for custom create action
	create    func(Entity) error
	initEmpty func() (Entity, error)
	findAll   iter.Seq2[Entity, error]
	delete    func(Entity) error // this is already an aggregate action
	//aggregateActions []AggregateAction[E]
	//binding          *Binding[E]
	wnd core.Window
}
