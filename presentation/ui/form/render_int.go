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

func RenderInt(ctx FieldContext) core.View {
	switch ctx.Field().Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	default:
		return nil
	}

	requiresInit := false
	intState := core.DerivedState[int64](ctx.State(), ctx.Field().Name).Init(func() int64 {
		src := ctx.State().Get()
		v := reflect.ValueOf(src).FieldByName(ctx.Field().Name).Int()
		if val := ctx.Value(); val != "" && v == 0 {
			p, err := strconv.ParseInt(val, 10, 64)
			if err == nil {
				requiresInit = true
				return p
			}
		}

		return v
	})

	intState.Observe(func(newValue int64) {
		ctx.SetValue(newValue)
	})

	if requiresInit {
		intState.Notify()
	}

	return ui.IntField(ctx.Label(), intState.Get(), intState).
		Disabled(ctx.Disabled()).
		SupportingText(ctx.SupportingText()).
		ErrorText(ctx.ErrorText()).
		Frame(ui.Frame{}.FullWidth())

}
