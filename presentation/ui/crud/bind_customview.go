package crud

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// CustomView creates a simple delegate field rendering for the given binding and state.
// Probably, you should always use this as a last resort, if no other already defined binding is applicable.
// You will miss validation, future design updates and lifecycle updates.
func CustomView[E any](makeView func(entity *core.State[E]) ui.DecoredView) Field[E] {
	return Field[E]{
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			return makeView(entity)
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			v := makeView(entity)
			return ui.TableCell(v)
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			return makeView(entity)
		},
		Comparator: nil,
		Stringer: func(e E) string {
			return fmt.Sprintf("%v", e)
		},
	}
}
