// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/picker"
	"reflect"
	"strings"
)

type PickMultipleOptions[T any] struct {
	Label  string
	Values []T
}

// PickMultiple binds a slice of an arbitrary value type to an associate picker. To pick an entity based
// on a foreign key semantics, use [OneToMany].
func PickMultiple[E any, T any](opts PickMultipleOptions[T], property Property[E, []T]) Field[E] {
	return Field[E]{
		Label: opts.Label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[[]T](self.Window, self.ID+"-form.local").Init(func() []T {
				var tmp E
				tmp = entity.Get()
				return property.Get(&tmp)
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue []T) {
				var tmp E
				tmp = entity.Get()
				oldValue := property.Get(&tmp)
				property.Set(&tmp, newValue)
				entity.Set(tmp)

				if !reflect.DeepEqual(oldValue, newValue) {
					entity.Notify()
				}

				handleValidation(self, entity, errState)
			})

			if self.requiresValidation() {
				state.Notify()
			}

			return picker.Picker[T](opts.Label, opts.Values, state).
				Title(self.Label).
				MultiSelect(true).
				Disabled(self.Disabled).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := property.Get(&tmp)
			return ui.TableCell(ui.Text(fmtSlice(v)))
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
