package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/picker"
	"iter"
	"log/slog"
	"strings"
)

type OneToManyTableOptions[T data.Aggregate[IDOfT], IDOfT data.IDType] struct {
	// Label of the field
	Label string
	// ForeignEntities contains the sequence of all entities which must be referenced through IDOfT.
	// The current implementation loads the entire set into memory, thus keep that number as small as possible.
	ForeignEntities iter.Seq2[T, error]
	// ForeignBinding provides the binding configuration to display the associated table of all types.
	ForeignBinding *Binding[T]
	// ForeignZero value is passed before rendering into the entity creation dialog
	ForeignZero T
	// ForeignCreate function to be invoked after ForeignBinding validation passed.
	ForeignCreate func(T) (errorText string, infrastructureError error)
	// ForeignPickerRenderer converts a T into a View for the picker dialog step. If nil, the value is
	// transformed using %v into a TextView.
	ForeignPickerRenderer func(T) core.View
}

// OneToManyTable binds a field with foreign key characteristics to a picker. See also [PickMultiple] for value semantics and [OneToMany] for a compact selection.
func OneToManyTable[E any, T data.Aggregate[IDOfT], IDOfT data.IDType](opts OneToManyTableOptions[T, IDOfT], property Property[E, []IDOfT]) Field[E] {
	if opts.ForeignPickerRenderer == nil {
		opts.ForeignPickerRenderer = func(t T) core.View {
			return ui.Text(fmt.Sprintf("%v", t))
		}
	}

	var values []T
	for v, err := range opts.ForeignEntities {
		if err != nil {
			slog.Error("OneToManyTable cannot get entity from Seq2, value is ignored", "err", err)
			continue
		}

		values = append(values, v)
	}

	valuesLookupById := map[IDOfT]T{}
	for _, fkEntity := range values {
		valuesLookupById[fkEntity.Identity()] = fkEntity
	}

	return Field[E]{
		Label: opts.Label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[[]T](self.Window, self.ID+"-form.local").Init(func() []T {
				var tmp E
				tmp = entity.Get()
				ids := property.Get(&tmp)

				resolvedEntities := make([]T, 0, len(ids))
				for _, id := range ids {
					// ignore orphaned
					v, ok := valuesLookupById[id]
					if ok {
						resolvedEntities = append(resolvedEntities, v)
					}
				}

				return resolvedEntities
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue []T) {
				var tmp E
				tmp = entity.Get()

				ids := make([]IDOfT, 0, len(newValue))
				for _, t := range newValue {
					ids = append(ids, t.Identity())
				}

				property.Set(&tmp, ids)
				entity.Set(tmp)

				handleValidation(self, entity, errState)
			})

			if self.requiresValidation() {
				state.Notify()
			}

			var pcker picker.TPicker[T]

			createTDlg, tDlgPresented := CreateDialog(opts.ForeignBinding, opts.ForeignZero, opts.ForeignCreate, func() {
				// on cancel
				pcker.DialogPresented().Set(true)
			}, func() {
				// on save
				pcker.DialogPresented().Set(true)
			})

			// this is always rendered and thus does not reset its state properly
			pcker = picker.Picker[T](opts.Label, values, state).
				Title(self.Label).
				ItemRenderer(func(t T) core.View {
					return opts.ForeignPickerRenderer(t)
				}).
				DetailView(ui.VStack(
					ui.HLine(),
					ui.TertiaryButton(func() {
						pcker.DialogPresented().Set(false)
						tDlgPresented.Set(true)

					}).Title("Neu anlegen").PreIcon(heroOutline.PlusCircle),
				).Alignment(ui.Leading).Frame(ui.Frame{}.FullWidth())).
				MultiSelect(true)

			pickerVisible := pcker.DialogPresented()

			selectedTIt := xslices.ValuesWithError(state.Get(), nil)

			return ui.VStack(
				createTDlg,
				ui.If(pickerVisible.Get(), ui.Composable(pcker.Dialog)),
				ui.HStack(
					ui.VStack(
						ui.Text(opts.Label).Font(ui.Title),
						ui.HLineWithColor(ui.ColorAccent),
					),
					ui.Spacer(),
					ui.SecondaryButton(func() {
						pickerVisible.Set(true)
					}).PreIcon(heroSolid.Plus).
						Title(fmt.Sprintf("%s zuteilen", opts.Label)),
				).Frame(ui.Frame{}.FullWidth()),
				Table[T, IDOfT](Options[T, IDOfT](opts.ForeignBinding).
					FindAll(selectedTIt)).
					Frame(ui.Frame{}.FullWidth()),
			).Frame(ui.Frame{}.FullWidth())

		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := property.Get(&tmp)
			views := make([]core.View, 0, len(v))
			for _, t := range v {
				entity, ok := valuesLookupById[t]
				if !ok {
					slog.Error("OneToMany cannot reverse lookup id", "id", t)
					continue
				}

				views = append(views, opts.ForeignPickerRenderer(entity))
			}
			return ui.TableCell(ui.HStack(views...).Alignment(ui.Leading).Wrap(true).Gap(ui.L8))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := property.Get(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(fmtSlice(v)),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := property.Get(&a)
			bv := property.Get(&b)
			return strings.Compare(fmtSlice(av), fmtSlice(bv))
		},
		Stringer: func(e E) string {
			return fmtSlice(property.Get(&e))
		},
	}
}
