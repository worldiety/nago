package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/list"
	"slices"
)

func List[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) ui.DecoredView {
	ds := opts.datasource()
	bnd := opts.bnd

	var count int
	return ui.VStack(
		list.List(ui.Each(slices.Values(ds.List()), func(entity Entity) core.View {
			count++
			if bnd.renderListEntry == nil {
				return list.Entry().Headline(fmt.Sprint(entity))
			}

			return bnd.renderListEntry(entity)
		})...).Caption(ui.Text("Alle Einträge")).
			Footer(ui.Text(fmt.Sprintf("%d von %d Einträgen", count, count))).
			Frame(ui.Frame{}.FullWidth()),
	).FullWidth()

}
