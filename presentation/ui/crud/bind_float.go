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

type FloatOptions struct {
	Label string
}

// Float binds a property to a float field.
func Float[E any, T std.Float](opts FloatOptions, property Property[E, T]) Field[E] {
	return Field[E]{
		Label: opts.Label,
		RenderFormElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			// here we create a copy for the local form field
			state := core.StateOf[float64](self.Window, self.ID+"-form.local").Init(func() float64 {
				var tmp E
				tmp = entity.Get()
				return float64(property.Get(&tmp))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			// if the local field changes, we push our stuff into the given state (whatever that is)
			state.Observe(func(newValue float64) {
				var tmp E
				tmp = entity.Get()
				oldValue := property.Get(&tmp)
				property.Set(&tmp, T(newValue))
				entity.Set(tmp)
				if float64(oldValue) != newValue {
					entity.Notify()
				}

				handleValidation(self, entity, errState)
			})

			entity.Observe(func(newValue E) {
				tmp := entity.Get()
				v := float64(property.Get(&tmp))
				state.Set(v)
				state.Notify()
			})

			return ui.FloatField(opts.Label, state.Get(), state).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get()).
				Disabled(self.Disabled).
				Frame(ui.Frame{}.FullWidth())
		},
		RenderTableCell: func(self Field[E], entity *core.State[E]) ui.TTableCell {
			tmp := entity.Get()
			v := property.Get(&tmp)
			return ui.TableCell(ui.Text(fmtFloat(float64(v))))
		},
		RenderCardElement: func(self Field[E], entity *core.State[E]) ui.DecoredView {
			var tmp E
			tmp = entity.Get()
			v := property.Get(&tmp)
			return ui.VStack(
				ui.VStack(ui.Text(self.Label).Font(ui.SubTitle)).
					Alignment(ui.Leading).
					Frame(ui.Frame{}.FullWidth()),
				ui.Text(fmtFloat(float64(v))),
			).Alignment(ui.Trailing)
		},
		Comparator: func(a, b E) int {
			av := property.Get(&a)
			bv := property.Get(&b)
			d := av - bv
			if d < 0 {
				return -1
			}

			if d > 0 {
				return 1
			}

			return 0
		},
		Stringer: func(e E) string {
			return fmtFloat(float64(property.Get(&e)))
		},
	}
}

func fmtFloat(v float64) string {
	//return fmt.Sprintf("%f", v)
	return strconv.FormatFloat(v, 'f', -1, 64)
}
