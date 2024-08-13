package crud

import (
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// Card creates a card view based on the field bindings. The given value is mapped automatically, based on the binding.
// A Card is usually readonly.
func Card[T any](bnd *Binding[T], value *core.State[T]) ui.DecoredView {
	return ui.VStack(
		slices.Collect(func(yield func(view core.View) bool) {
			for _, field := range bnd.fields {
				if field.RenderCardElement != nil {
					yield(field.RenderCardElement(field, value).Frame(ui.Frame{}.FullWidth()))
				}
			}

		})...,
	).Gap(ui.L16).
		BackgroundColor(ui.M4).
		Padding(ui.Padding{}.All(ui.L8)).
		Border(ui.Border{}.Elevate(4))
}
