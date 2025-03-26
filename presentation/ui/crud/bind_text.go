// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"strings"
)

type TextOptions struct {
	Label           string
	KeyboardOptions ui.TKeyboardOptions
	Lines           int
	SupportingText  string
}

func Text[E any, T ~string](opts TextOptions, property Property[E, T]) Field[E] {
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
				oldValue := property.Get(&tmp)
				property.Set(&tmp, T(newValue))
				entity.Set(tmp)
				if string(oldValue) != newValue {
					entity.Notify()
				}

				handleValidation(self, entity, errState)
			})

			entity.Observe(func(newValue E) {
				tmp := entity.Get()
				v := string(property.Get(&tmp))
				state.Set(v)
				state.Notify()
			})

			if self.requiresValidation() {
				state.Notify()
			}

			return ui.TextField(opts.Label, state.String()).
				InputValue(state).
				Lines(opts.Lines).
				KeyboardOptions(opts.KeyboardOptions).
				Disabled(self.Disabled).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := property.Get(&tmp)
			return ui.TableCell(ui.Text(string(v)))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := property.Get(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(string(v)),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := property.Get(&a)
			bv := property.Get(&b)
			return strings.Compare(string(av), string(bv))
		},
		Stringer: func(e E) string {
			return string(property.Get(&e))
		},
	}
}
