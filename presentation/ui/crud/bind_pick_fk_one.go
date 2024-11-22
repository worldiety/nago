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

type OneToOneOptions[T data.Aggregate[IDOfT], IDOfT data.IDType] struct {
	Label string
	// ForeignEntities contains the sequence of all entities which must be referenced through IDOfT.
	// The current implementation loads the entire set into memory, thus keep that number as small as possible.
	ForeignEntities iter.Seq2[T, error]
	// ForeignPickerRenderer converts a T into a View for the picker dialog step. If nil, the value is
	// transformed using %v into a TextView.
	ForeignPickerRenderer func(T) core.View
}

// OneToOne binds a field with foreign key characteristics to a picker. See also [PickOne] for value
// semantics.
func OneToOne[E any, T data.Aggregate[IDOfT], IDOfT data.IDType](opts OneToOneOptions[T, IDOfT], property Property[E, std.Option[IDOfT]]) Field[E] {
	if opts.ForeignPickerRenderer == nil {
		opts.ForeignPickerRenderer = func(t T) core.View {
			return ui.Text(fmt.Sprintf("%v", t))
		}
	}

	var values []T
	for v, err := range opts.ForeignEntities {
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

	var zero IDOfT

	return Field[E]{
		Label: opts.Label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[[]T](self.Window, self.ID+"-form.local").Init(func() []T {
				var tmp E
				tmp = entity.Get()
				optId := property.Get(&tmp)

				if optId.IsSome() {
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

				if len(newValue) > 0 {
					property.Set(&tmp, std.Some[IDOfT](newValue[0].Identity()))
				} else {
					property.Set(&tmp, std.None[IDOfT]())
				}

				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			if self.requiresValidation() {
				state.Notify()
			}

			return picker.Picker[T](opts.Label, values, state).
				Title(self.Label).
				ItemRenderer(func(t T) core.View {
					return opts.ForeignPickerRenderer(t)
				}).
				ItemPickedRenderer(func(t []T) core.View {
					switch len(t) {
					case 0:
						return ui.Text("nichts gewählt")
					case 1:
						return opts.ForeignPickerRenderer(t[0])
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
			v := property.Get(&tmp)
			if v.IsSome() {
				ent, ok := valuesLookupById[v.Unwrap()]
				if !ok {
					slog.Error("OneToOne cannot reverse lookup id", "id", v.V)
					return ui.TableCell(ui.Text(""))
				}
				return ui.TableCell(opts.ForeignPickerRenderer(ent))
			}

			return ui.TableCell(nil)
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := property.Get(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				opts.ForeignPickerRenderer(valuesLookupById[v.V]),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {

			av := valuesLookupById[(property.Get(&a)).UnwrapOr(zero)]
			bv := valuesLookupById[(property.Get(&b)).UnwrapOr(zero)]
			return strings.Compare(fmt.Sprintf("%v", av), fmt.Sprintf("%v", bv))
		},
		Stringer: func(e E) string {
			return fmt.Sprintf("%v", property.Get(&e).UnwrapOr(zero))
		},
	}
}
