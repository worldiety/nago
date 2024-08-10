package crud

import (
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// Form creates a form view based on the field bindings. The given value is mapped automatically, based on the binding.
// The implementation pushes automatically
func Form[T any](bnd *Binding[T], value *T) ui.DecoredView {
	return ui.VStack(
		slices.Collect(func(yield func(view core.View) bool) {
			for _, field := range bnd.fields {
				if field.RenderFormElement != nil {
					yield(field.RenderFormElement(field, value))
				}
			}

		})...,
	)
}
