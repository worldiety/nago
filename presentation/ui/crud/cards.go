package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"slices"
)

func Cards[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) ui.DecoredView {
	ds := opts.datasource()
	bnd := opts.bnd

	return ui.VStack(
		ui.Each(slices.Values(ds.List()), func(entity Entity) core.View {
			entityState := core.StateOf[Entity](opts.wnd, fmt.Sprintf("crud.card.entity.%v", entity.Identity())).From(func() Entity {
				return entity
			})
			return Card[Entity](bnd, entityState).Frame(ui.Frame{}.FullWidth())
		})...,
	).Gap(ui.L16).Padding(ui.Padding{}.All(ui.L16))
}
