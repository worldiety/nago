package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func NewView[Entity data.Aggregate[ID], ID data.IDType](wnd core.Window, opts *Options[Entity, ID], bnd *Binding[Entity]) ui.DecoredView {
	ds := dataSource[Entity, ID]{
		it:          opts.findAll,
		binding:     bnd,
		sortByField: nil,
		query:       "",
	}
	opts.wnd = wnd
	return Table(opts, bnd, ds)
}
