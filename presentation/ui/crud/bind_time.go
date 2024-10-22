package crud

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/timepicker"
	"time"
)

// TimeOptions aggregates all available Options to bind a time picker. In the future, additional options may be added.
type TimeOptions struct {
	Label         string
	ShowDays      bool
	ShowHours     bool
	ShowMinutes   bool
	ShowSeconds   bool
	DisplayFormat timepicker.PickerFormat
}

func Time[E any](opts TimeOptions, property Property[E, time.Duration]) Field[E] {
	if opts.ShowDays == false && opts.ShowHours == false && opts.ShowMinutes == false && opts.ShowSeconds == false {
		// provide some sane default behavior
		opts.ShowHours = true
		opts.ShowMinutes = true
	}

	formatTime := func(entity E) string {
		return timepicker.Format(opts.ShowDays, opts.ShowHours, opts.ShowMinutes, opts.ShowSeconds, opts.DisplayFormat, property.Get(&entity))
	}

	return Field[E]{
		Label: opts.Label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[time.Duration](self.Window, self.ID+"-form.local").Init(func() time.Duration {
				var tmp E
				tmp = entity.Get()
				return property.Get(&tmp)
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue time.Duration) {
				var tmp E
				tmp = entity.Get()
				property.Set(&tmp, newValue)
				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			if self.requiresValidation() {
				state.Notify()
			}

			return timepicker.Picker(opts.Label, state).
				Format(opts.DisplayFormat).
				Days(opts.ShowDays).
				Hours(opts.ShowHours).
				Minutes(opts.ShowMinutes).
				Seconds(opts.ShowSeconds).
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
			av := property.Get(&a)
			bv := property.Get(&b)
			return int(av - bv)
		},
		Stringer: func(e E) string {
			return formatTime(e)
		},
	}
}
