package crud

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/picker"
	"strings"
)

// PickMultiple binds a slice of an arbitrary value type to an associate picker. To pick an entity based
// on a foreign key semantics, use [OneToMany].
func PickMultiple[E any, T any](label string, values []T, property func(model *E) *[]T) Field[E] {
	return Field[E]{
		Label: label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[[]T](self.Window, self.ID+"-form.local").From(func() []T {
				var tmp E
				tmp = entity.Get()
				return *property(&tmp)
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue []T) {
				var tmp E
				tmp = entity.Get()
				f := property(&tmp)
				*f = newValue
				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			return picker.Picker[T](label, values, state).
				Title(self.Label).
				MultiSelect(true).
				Disabled(self.Disabled).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := *property(&tmp)
			return ui.TableCell(ui.Text(fmtSlice(v)))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := *property(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(fmtSlice(v)),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := *property(&a)
			bv := *property(&b)
			return strings.Compare(fmtSlice(av), fmtSlice(bv))
		},
		Stringer: func(e E) string {
			return fmtSlice(*property(&e))
		},
	}
}

func fmtSlice[T any](v []T) string {
	sb := strings.Builder{}
	for i, t := range v {
		sb.WriteString(fmt.Sprintf("%v", t))
		if i < len(v)-1 {
			sb.WriteString(", ")
		}
	}

	return sb.String()
}
