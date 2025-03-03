package crud

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"strconv"
)

type BoolOptions struct {
	Label          string
	SupportingText string
}

func Bool[E any, T ~bool](opts BoolOptions, property Property[E, T]) Field[E] {
	return Field[E]{
		Label:          opts.Label,
		SupportingText: opts.SupportingText,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[bool](self.Window, self.ID+"-form.local").Init(func() bool {
				var tmp E
				tmp = entity.Get()
				return bool(property.Get(&tmp))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue bool) {
				var tmp E
				tmp = entity.Get()
				property.Set(&tmp, T(newValue))
				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			return ui.CheckboxField(opts.Label, state.Get()).
				InputValue(state).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
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
			if av && !bv {
				return 1
			}

			if !av && bv {
				return -1
			}

			return 0
		},
		Stringer: func(e E) string {
			v := bool(property.Get(&e))
			if v {
				return "ja"
			} else {
				return "nein"
			}
		},
	}
}

func BoolToggle[E any, T ~bool](opts BoolOptions, property Property[E, T]) Field[E] {
	return Field[E]{
		Label: opts.Label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[bool](self.Window, self.ID+"-form.local").Init(func() bool {
				var tmp E
				tmp = entity.Get()
				return bool(property.Get(&tmp))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue bool) {
				var tmp E
				tmp = entity.Get()
				property.Set(&tmp, T(newValue))
				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			return ui.ToggleField(opts.Label, state.Get()).
				InputValue(state).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := property.Get(&tmp)
			return ui.TableCell(ui.Text(strconv.FormatBool(bool(v))))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := property.Get(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(strconv.FormatBool(bool(v))),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := property.Get(&a)
			bv := property.Get(&b)
			if av && !bv {
				return 1
			}

			if !av && bv {
				return -1
			}

			return 0
		},
		Stringer: func(e E) string {
			return strconv.FormatBool(bool(property.Get(&e)))
		},
	}
}
