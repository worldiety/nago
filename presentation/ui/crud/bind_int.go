package crud

import (
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"strconv"
)

func Int[E any, T std.Integer](label string, property func(model *E) *T) Field[E] {
	return Field[E]{
		Label: label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[int64](self.Window, self.ID+"-form.local").From(func() int64 {
				var tmp E
				tmp = entity.Get()
				return int64(*property(&tmp))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue int64) {
				var tmp E
				tmp = entity.Get()
				f := property(&tmp)
				*f = T(newValue)
				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			return ui.IntField(label, state.Get(), state).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := *property(&tmp)
			return ui.TableCell(ui.Text(strconv.FormatInt(int64(v), 10)))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := *property(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(strconv.FormatInt(int64(v), 10)),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := *property(&a)
			bv := *property(&b)
			return int(av - bv)
		},
		Stringer: func(e E) string {
			return strconv.FormatInt(int64(*property(&e)), 10)
		},
	}
}
