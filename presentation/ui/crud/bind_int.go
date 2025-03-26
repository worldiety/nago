// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"strconv"
)

type IntOptions struct {
	Label          string
	SupportingText string
}

func Int[E any, T std.Integer](opts IntOptions, property Property[E, T]) Field[E] {
	return Field[E]{
		Label:          opts.Label,
		SupportingText: opts.SupportingText,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[int64](self.Window, self.ID+"-form.local").Init(func() int64 {
				var tmp E
				tmp = entity.Get()
				return int64(property.Get(&tmp))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue int64) {
				var tmp E
				tmp = entity.Get()
				oldValue := property.Get(&tmp)
				property.Set(&tmp, T(newValue))
				entity.Set(tmp)
				if int64(oldValue) != newValue {
					entity.Notify()
				}

				handleValidation(self, entity, errState)
			})

			entity.Observe(func(newValue E) {
				tmp := entity.Get()
				v := int64(property.Get(&tmp))
				state.Set(v)
				state.Notify()
			})

			if self.requiresValidation() {
				state.Notify()
			}

			return ui.IntField(opts.Label, state.Get(), state).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := property.Get(&tmp)
			return ui.TableCell(ui.Text(strconv.FormatInt(int64(v), 10)))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := property.Get(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(strconv.FormatInt(int64(v), 10)),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := property.Get(&a)
			bv := property.Get(&b)
			return int(av - bv)
		},
		Stringer: func(e E) string {
			return strconv.FormatInt(int64(property.Get(&e)), 10)
		},
	}
}
