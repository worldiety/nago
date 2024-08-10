package crud

import (
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// Views creates a field binding to E and renders with the binded E the given options.
// Keep in mind, to remove Render* functions, if it does not make sense or may cause
// malfunctions in the context, e.g. deleting an E without navigation.
func Views[E any](label string, options ...func(*E) core.View) Field[E] {
	return Field[E]{
		Label: label,
		RenderTableCell: func(self Field[E], entity *E) ui.TTableCell {
			return ui.TableCell(ui.HStack(slices.Collect(func(yield func(cell core.View) bool) {
				for _, option := range options {
					yield(option(entity))
				}
			})...).
				Gap(ui.L4).
				Alignment(ui.Trailing))
		},
		RenderCardElement: func(self Field[E], entity *E) ui.DecoredView {
			return ui.HStack(slices.Collect(func(yield func(cell core.View) bool) {
				for _, option := range options {
					yield(option(entity))
				}
			})...).
				Gap(ui.L4).
				Alignment(ui.Leading)
		},

		RenderFormElement: func(self Field[E], entity *E) ui.DecoredView {
			return self.RenderCardElement(self, entity)
		},
	}
}
