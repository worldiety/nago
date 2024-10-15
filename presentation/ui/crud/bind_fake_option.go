package crud

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func Option[E any](label string, enabled bool, fields ...Field[E]) []Field[E] {
	return fakeFormFields("", optionSection(fields, label, enabled), fields...)
}

func optionSection[E any](fields []Field[E], label string, enabled bool) func(bnd *Binding[E], views ...core.View) ui.DecoredView {
	return func(bnd *Binding[E], views ...core.View) ui.DecoredView {
		checkedState := core.StateOf[bool](bnd.wnd, bnd.id+label).Init(func() bool {
			return enabled
		})
		cb := ui.Checkbox(checkedState.Get()).InputChecked(checkedState)

		allViews := make([]core.View, 0, len(fields)+1)
		allViews = append(allViews, ui.HStack(
			cb,
			ui.Text(label),
		))

		if checkedState.Get() {
			allViews = append(allViews, views...)
		}

		return ui.VStack(allViews...).
			Alignment(ui.Leading).
			Frame(ui.Frame{}.FullWidth())
	}
}
