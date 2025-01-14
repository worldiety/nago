package crud

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type PasswordOptions struct {
	Label          string
	SupportingText string
}

func Password[E any, T ~string](opts PasswordOptions, property Property[E, T]) Field[E] {
	return Field[E]{
		Label:          opts.Label,
		SupportingText: opts.SupportingText,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[string](self.Window, self.ID+"-form.local").Init(func() string {
				var tmp E
				tmp = entity.Get()
				return string(property.Get(&tmp))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue string) {
				var tmp E
				tmp = entity.Get()
				property.Set(&tmp, T(newValue))
				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			if self.requiresValidation() {
				state.Notify()
			}

			return ui.PasswordField(opts.Label, state.Get()).
				InputValue(state).
				Disabled(self.Disabled).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			return ui.TableCell(ui.Text("****"))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text("****"),
			).Alignment(ui.Trailing)
		},
		Comparator: nil,
		Stringer: func(e E) string {
			return "****"
		},
	}
}
