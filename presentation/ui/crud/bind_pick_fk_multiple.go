// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"fmt"
	"iter"
	"log/slog"
	"reflect"
	"strings"

	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/picker"
)

type OneToManyOptions[T data.Aggregate[IDOfT], IDOfT data.IDType] struct {
	Label string
	// ForeignEntities contains the sequence of all entities which must be referenced through IDOfT.
	// The current implementation loads the entire set into memory, thus keep that number as small as possible.
	ForeignEntities iter.Seq2[T, error]
	// ForeignPickerRenderer converts a T into a View for the picker dialog step. If nil, the value is
	// transformed using %v into a TextView.
	ForeignPickerRenderer func(T) core.View

	SupportingText string
}

// OneToMany binds a field with foreign key characteristics to a picker. See also [PickMultiple] for value
// semantics.
func OneToMany[E any, T data.Aggregate[IDOfT], IDOfT data.IDType](opts OneToManyOptions[T, IDOfT], property Property[E, []IDOfT]) Field[E] {
	if opts.ForeignPickerRenderer == nil {
		opts.ForeignPickerRenderer = func(t T) core.View {
			return ui.Text(fmt.Sprintf("%v", t)).Resolve(true)
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

	return Field[E]{
		Label:          opts.Label,
		SupportingText: opts.SupportingText,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[[]T](self.Window, self.ID+"-form.local").Init(func() []T {
				var tmp E
				tmp = entity.Get()
				ids := property.Get(&tmp)

				resolvedEntities := make([]T, 0, len(ids))
				for _, id := range ids {
					// id may be orphaned
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
				oldValue := property.Get(&tmp)

				ids := make([]IDOfT, 0, len(newValue))
				for _, t := range newValue {
					ids = append(ids, t.Identity())
				}

				property.Set(&tmp, ids)
				entity.Set(tmp)

				if !reflect.DeepEqual(oldValue, newValue) {
					entity.Notify()
				}

				handleValidation(self, entity, errState)
			})

			if self.requiresValidation() {
				state.Notify()
			}

			textColor := ui.Color("")
			if self.Disabled {
				textColor = "ST0"
			}
			return picker.Picker[T](opts.Label, values, state).
				Title(self.Label).
				ItemRenderer(func(t T) core.View {
					return opts.ForeignPickerRenderer(t)
				}).
				ItemPickedRenderer(func(t []T) core.View {
					switch len(t) {
					case 0:
						return ui.Text("nichts gewählt").Color(textColor)
					case 1:
						return opts.ForeignPickerRenderer(t[0])
					default:
						return ui.Text(fmt.Sprintf("%d gewählt", len(t))).Color(textColor)
					}
				}).
				MultiSelect(true).
				Disabled(self.Disabled).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Frame(ui.Frame{}.FullWidth())
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
