package crud

import (
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/timeframe"
	"time"
)

type TimeFrameOptions struct {
	Label    string
	Location *time.Location // default is UTC
}

func TimeFrame[E any, T ~struct {
	StartTime xtime.UnixMilliseconds // inclusive
	EndTime   xtime.UnixMilliseconds // inclusive
	Timezone  xtime.Timezone
}](opts TimeFrameOptions, property Property[E, T]) Field[E] {
	if opts.Location == nil {
		opts.Location = time.UTC
	}

	return Field[E]{
		Label: opts.Label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[xtime.TimeFrame](self.Window, self.ID+"-form.local").Init(func() xtime.TimeFrame {
				var tmp E
				tmp = entity.Get()
				tf := xtime.TimeFrame(property.Get(&tmp))
				if tf.Timezone == "" {
					tf.Timezone = xtime.Timezone(opts.Location.String())
				}

				return xtime.TimeFrame(tf)
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue xtime.TimeFrame) {
				var tmp E
				tmp = entity.Get()
				property.Set(&tmp, T(newValue))
				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			return timeframe.Picker(opts.Label, state).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Disabled(self.Disabled).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			return ui.TableCell(ui.Text(self.Stringer(tmp)))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(self.Stringer(tmp)),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := property.Get(&a)
			bv := property.Get(&b)
			if av == bv {
				return 0
			}

			if xtime.TimeFrame(av).StartTime > xtime.TimeFrame(bv).StartTime {
				return 1
			} else {
				return -1
			}
		},
		Stringer: func(e E) string {
			val := property.Get(&e)
			if xtime.TimeFrame(val).IsZero() {
				return ""
			}

			return xtime.TimeFrame(val).Format(xtime.GermanDate)
		},
	}
}
