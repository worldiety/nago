package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/picker"
	"strings"
)

// PickOne binds a single field of an arbitrary value type to an associate picker. To pick an entity based
// on a foreign key semantics, use [OneToOne].
func PickOne[E any, T any](label string, values []T, property func(model *E) *std.Option[T]) Field[E] {
	return Field[E]{
		Label: label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[[]T](self.Window, self.ID+"-form.local").From(func() []T {
				var tmp E
				tmp = entity.Get()
				optT := *property(&tmp)
				if optT.Valid {
					return []T{optT.Unwrap()}
				}

				return []T{}
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue []T) {
				var tmp E
				tmp = entity.Get()
				f := property(&tmp)
				if len(newValue) == 0 {
					*f = std.None[T]()
				} else {
					*f = std.Some[T](newValue[0])
				}

				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			if self.requiresValidation() {
				state.Notify()
			}

			return picker.Picker[T](label, values, state).
				Title(self.Label).
				MultiSelect(false).
				Disabled(self.Disabled).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := *property(&tmp)
			return ui.TableCell(ui.Text(fmtOptOne(v)))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := *property(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(fmtOptOne(v)),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := *property(&a)
			bv := *property(&b)
			return strings.Compare(fmtOptOne(av), fmtOptOne(bv))
		},
		Stringer: func(e E) string {
			return fmtOptOne(*property(&e))
		},
	}
}

func fmtOptOne[T any](v std.Option[T]) string {
	if v.Valid {
		return fmt.Sprintf("%v", v.Unwrap())
	}

	return ""
}
