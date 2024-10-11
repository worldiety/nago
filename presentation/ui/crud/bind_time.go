package crud

import (
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/timepicker"
)

func Time[E any, T std.Integer](label string, scaleToSeconds int64, days, hours, minutes, seconds bool, format timepicker.PickerFormat, property func(model *E) *T) Field[E] {
	formatTime := func(entity E) string {
		return timepicker.Format[T](scaleToSeconds, days, hours, minutes, seconds, format, *property(&entity))
	}

	return Field[E]{
		Label: label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[T](self.Window, self.ID+"-form.local").From(func() T {
				var tmp E
				tmp = entity.Get()
				return T(*property(&tmp))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue T) {
				var tmp E
				tmp = entity.Get()
				f := property(&tmp)
				*f = T(newValue)
				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			if self.requiresValidation() {
				state.Notify()
			}

			return timepicker.Picker[T](label, scaleToSeconds, state).
				Format(format).
				Days(days).
				Hours(hours).
				Minutes(minutes).
				Seconds(seconds).
				Disabled(self.Disabled).
				ErrorText(errState.Get()).
				SupportingText(self.SupportingText).
				Frame(ui.Frame{}.FullWidth())

		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			return ui.TableCell(ui.Text(formatTime(tmp)))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(formatTime(tmp)),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := *property(&a)
			bv := *property(&b)
			return int(av - bv)
		},
		Stringer: func(e E) string {
			return formatTime(e)
		},
	}
}
