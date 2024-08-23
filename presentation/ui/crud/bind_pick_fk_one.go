package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/picker"
	"iter"
	"log/slog"
	"strings"
)

// OneToOne binds a field with foreign key characteristics to a picker. See also [PickOne] for value
// semantics.
func OneToOne[E any, T data.Aggregate[IDOfT], IDOfT data.IDType](label string, it iter.Seq2[T, error], fkRenderer func(T) core.View, property func(model *E) *std.Option[IDOfT]) Field[E] {
	var values []T
	for v, err := range it {
		if err != nil {
			slog.Error("OneToMany cannot get entity from Seq2, value is ignored", "err", err)
			continue
		}

		values = append(values, v)
	}

	valuesLookupById := map[IDOfT]T{}
	for _, fkEntity := range values {
		valuesLookupById[fkEntity.Identity()] = fkEntity
	}

	return Field[E]{
		Label: label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[[]T](self.Window, self.ID+"-form.local").From(func() []T {
				var tmp E
				tmp = entity.Get()
				optId := *property(&tmp)

				if optId.Valid {
					ent, ok := valuesLookupById[optId.Unwrap()]
					if ok {
						return []T{ent}
					} else {
						slog.Error("OneToOne cannot lookup selected entry", "id", optId.V)
					}
				}

				return nil
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue []T) {
				var tmp E
				tmp = entity.Get()
				f := property(&tmp)

				if len(newValue) > 0 {
					*f = std.Some[IDOfT](newValue[0].Identity())
				} else {
					*f = std.None[IDOfT]()
				}

				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			return picker.Picker[T](label, values, state).
				Title(self.Label).
				ItemRenderer(func(t T) core.View {
					return fkRenderer(t)
				}).
				ItemPickedRenderer(func(t []T) core.View {
					switch len(t) {
					case 0:
						return ui.Text("nichts gewählt")
					case 1:
						return fkRenderer(t[0])
					default:
						return ui.Text(fmt.Sprintf("%d gewählt", len(t)))
					}
				}).
				MultiSelect(false).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Disabled(self.Disabled).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := *property(&tmp)
			if v.Valid {
				ent, ok := valuesLookupById[v.V]
				if !ok {
					slog.Error("OneToOne cannot reverse lookup id", "id", v.V)
					return ui.TableCell(ui.Text(""))
				}
				return ui.TableCell(fkRenderer(ent))
			}

			return ui.TableCell(nil)
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := *property(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				fkRenderer(valuesLookupById[v.V]),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := valuesLookupById[(*property(&a)).UnwrapOrZero()]
			bv := valuesLookupById[(*property(&b)).UnwrapOrZero()]
			return strings.Compare(fmt.Sprintf("%v", av), fmt.Sprintf("%v", bv))
		},
		Stringer: func(e E) string {
			return fmt.Sprintf("%v", (*property(&e)).UnwrapOrZero())
		},
	}
}
