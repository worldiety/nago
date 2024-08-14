package crud

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"slices"
)

// Form creates a form view based on the field bindings. The given value is mapped automatically, based on the binding.
// The implementation pushes automatically
func Form[T any](bnd *Binding[T], value *core.State[T]) ui.DecoredView {
	return ui.VStack(
		slices.Collect(func(yield func(view core.View) bool) {
			for _, field := range bnd.fields {
				if field.RenderFormElement != nil {
					yield(ui.Composable(func() core.View {
						return field.RenderFormElement(field, value)
					}))
				}
			}

		})...,
	).Gap(ui.L16).Frame(ui.Frame{}.FullWidth())
}
