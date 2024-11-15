package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/list"
	"slices"
)

func List[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) ui.DecoredView {
	ds := opts.datasource()
	//	bnd := opts.bnd

	return ui.VStack(
		list.List(ui.Each(slices.Values(ds.List()), func(entity Entity) core.View {
			/*slices.Values(bnd.listViewFields()
			entityState := core.StateOf[Entity](opts.wnd, fmt.Sprintf("crud.row.entity.%v", entity.Identity())).Init(func() Entity {
				return entity
			})
			return t.RenderListEntry(t)*/
			panic("fix me")
		})...),
	)

}
