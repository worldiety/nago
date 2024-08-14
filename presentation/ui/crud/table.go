package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"reflect"
	"slices"
)

func Table[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) ui.DecoredView {
	ds := opts.datasource()
	bnd := opts.bnd

	return ui.VStack(

		ui.Table(ui.Each(slices.Values(bnd.fields), func(field Field[Entity]) ui.TTableColumn {
			return ui.TableColumn(ui.IfElse(field.Comparator == nil, ui.Text(field.Label), ui.TertiaryButton(func() {

			}).Title(field.Label).Font(ui.Font{Size: ui.L16, Weight: ui.NormalFontWeight}))).Padding(ui.Padding{Left: ui.L0, Right: ui.L24, Top: ui.L16, Bottom: ui.L16})
		})...,
		).Rows(ui.Each(slices.Values(ds.List()), func(entity Entity) ui.TTableRow {
			entityState := core.StateOf[Entity](opts.wnd, fmt.Sprintf("crud.row.entity.%v", entity.Identity())).From(func() Entity {
				return entity
			})

			if !reflect.DeepEqual(entityState.Get(), entity) {
				entityState.Set(entity)
			}

			var cells []ui.TTableCell
			for _, field := range bnd.fields {
				if field.RenderTableCell == nil {
					cells = append(cells, ui.TableCell(nil))
				} else {
					cells = append(cells, field.RenderTableCell(field, entityState))
				}
			}

			return ui.TableRow(cells...)
		})...).Frame(ui.Frame{}.FullWidth()),
	)
}
