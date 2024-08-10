package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func Table[Entity data.Aggregate[ID], ID data.IDType](opts *Options[Entity, ID], bnd *Binding[Entity], ds dataSource[Entity, ID]) ui.DecoredView {
	quickSearch := core.AutoState[string](opts.wnd)

	return ui.VStack(
		ui.HStack(
			ui.TextField("", quickSearch.String()).InputValue(quickSearch).Style(ui.TextFieldReduced),
		),
		ui.Table(ui.Each(slices.Values(bnd.fields), func(field Field[Entity]) ui.TTableColumn {
			return ui.TableColumn(ui.IfElse(field.Comparator == nil, ui.Text(field.Label), ui.TertiaryButton(nil).Title(field.Label)))
		})...,
		).Rows(ui.Each(slices.Values(ds.List()), func(entity Entity) ui.TTableRow {
			var cells []ui.TTableCell
			for _, field := range bnd.fields {
				if field.RenderTableCell == nil {
					cells = append(cells, ui.TableCell(nil))
				} else {
					cells = append(cells, field.RenderTableCell(field, &entity))
				}
			}

			return ui.TableRow(cells...)
		})...),
	)
}
