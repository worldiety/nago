package crud

import (
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/colorpicker"
	"strings"
)

type PickOneColorOptions[T any] struct {
	Label   string
	Palette colorpicker.Palette // default is [colorpicker.DefaultPalette]
}

// PickOneColor binds a single field of a color value type to an associated picker. To pick an entity based
// on a foreign key semantics, use [OneToOne]. The T type follows the semantics of [ui.Color].
func PickOneColor[E any, T ~string](opts PickOneColorOptions[T], property Property[E, std.Option[T]]) Field[E] {
	if opts.Palette == nil {
		opts.Palette = colorpicker.DefaultPalette
	}

	return Field[E]{
		Label: opts.Label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[ui.Color](self.Window, self.ID+"-form.local").Init(func() ui.Color {
				var tmp E
				tmp = entity.Get()
				optT := property.Get(&tmp)
				return ui.Color(optT.UnwrapOr(""))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue ui.Color) {
				var tmp E
				tmp = entity.Get()
				if len(newValue) == 0 {
					property.Set(&tmp, std.None[T]())
				} else {
					property.Set(&tmp, std.Some[T](T(newValue)))
				}

				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			if self.requiresValidation() {
				state.Notify()
			}

			return colorpicker.PalettePicker(opts.Label, opts.Palette).
				Title(self.Label).
				Value(state.Get()).
				State(state).
				Disabled(self.Disabled).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := property.Get(&tmp)
			return ui.TableCell(colorpicker.Color(ui.Color(v.UnwrapOr(""))))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := property.Get(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				colorpicker.Color(ui.Color(v.UnwrapOr(""))),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := property.Get(&a)
			bv := property.Get(&b)
			return strings.Compare(fmtOptOne(av), fmtOptOne(bv))
		},
		Stringer: func(e E) string {
			return fmtOptOne(property.Get(&e))
		},
	}
}
