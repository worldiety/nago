package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func renderTable[Entity data.Aggregate[ID], ID data.IDType](opts *Options[Entity, ID], bnd *Binding[Entity, ID], ds dataSource[Entity, ID]) ui.DecoredView {
	quickSearch := core.AutoState[string](opts.wnd)
	return ui.VStack(
		ui.HStack(
			ui.TextField("", quickSearch.String()).InputValue(quickSearch).Style(ui.TextFieldReduced),
		),
		ui.Table(),
	)
}
