// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"time"
)

type DateOptions struct {
	Label string
}

func Date[E any, T ~struct {
	Day   int        // Year like 2024.
	Month time.Month // Month in year, offset at 1.
	Year  int        // Day of month, offset at 1.
}](opts DateOptions, property Property[E, T]) Field[E] {
	return Field[E]{
		Label: opts.Label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[xtime.Date](self.Window, self.ID+"-form.local").Init(func() xtime.Date {
				var tmp E
				tmp = entity.Get()

				return xtime.Date(property.Get(&tmp))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue xtime.Date) {
				var tmp E
				tmp = entity.Get()
				oldValue := property.Get(&tmp)
				property.Set(&tmp, T(newValue))
				entity.Set(tmp)

				if xtime.Date(oldValue) != newValue {
					entity.Notify()
				}

				handleValidation(self, entity, errState)
			})

			return ui.SingleDatePicker(opts.Label, state.Get(), state).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Disabled(self.Disabled).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			return ui.TableCell(ui.Text(self.Stringer(tmp)))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(self.Stringer(tmp)),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := property.Get(&a)
			bv := property.Get(&b)
			if av == bv {
				return 0
			}

			if xtime.Date(av).After(xtime.Date(bv)) {
				return 1
			} else {
				return -1
			}
		},
		Stringer: func(e E) string {
			val := property.Get(&e)
			if xtime.Date(val).Zero() {
				return ""
			}
			return xtime.Date(val).Format(xtime.GermanDate)
		},
	}
}
