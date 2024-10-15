package crud

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type FormCol[E any] struct {
	field  Field[E]
	weight core.Weight
}

func FormColumn[E any](field Field[E], weight core.Weight) FormCol[E] {
	return FormCol[E]{
		field:  field,
		weight: weight,
	}
}

// Row packs the given fields into a single row for the form. It works exactly like [Section] otherwise.
func Row[E any](cols ...FormCol[E]) []Field[E] {
	tmp := make([]Field[E], 0, len(cols))
	for _, col := range cols {
		tmp = append(tmp, col.field)
	}
	return fakeFormFields("", rowStack(cols...), tmp...)
}

func rowStack[E any](cols ...FormCol[E]) func(bnd *Binding[E], views ...core.View) ui.DecoredView {
	return func(bnd *Binding[E], views ...core.View) ui.DecoredView {
		cells := make([]ui.TGridCell, 0, len(cols))
		widths := make([]ui.Length, 0, len(cells))
		for idx, col := range cols {
			cells = append(cells, ui.GridCell(views[idx]))
			widths = append(widths, ui.Length(fmt.Sprintf("%dfr", int(col.weight*100))))
		}

		return ui.Grid(cells...).
			Rows(1).
			Columns(len(cols)).
			Widths(widths...).
			Gap(ui.L16).
			Frame(ui.Frame{}.FullWidth())
	}
}
