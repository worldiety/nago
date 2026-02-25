// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"reflect"
	"strconv"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func RenderFloat(ctx FieldContext) core.View {
	if ctx.Field().Type.Kind() != reflect.Float64 && ctx.Field().Type.Kind() != reflect.Float32 {
		return nil
	}

	requiresInit := false
	floatState := core.DerivedState[float64](ctx.State(), ctx.Field().Name).Init(func() float64 {
		src := ctx.State().Get()
		v := reflect.ValueOf(src).FieldByName(ctx.Field().Name).Float()
		if val := ctx.Value(); val != "" && v == 0 {
			p, err := strconv.ParseFloat(val, 64)
			if err == nil {
				requiresInit = true
				return p
			}
		}

		return v
	})

	floatState.Observe(func(newValue float64) {
		ctx.SetValue(newValue)
	})

	if requiresInit {
		floatState.Notify()
	}

	return ui.FloatField(ctx.Label(), floatState.Get(), floatState).
		Disabled(ctx.Disabled()).
		ID(ctx.ID()).
		SupportingText(ctx.SupportingText()).
		ErrorText(ctx.ErrorText()).
		FullWidth()

}
