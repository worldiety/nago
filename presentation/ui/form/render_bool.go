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

func RenderBool(ctx FieldContext) core.View {
	if ctx.Field().Type.Kind() != reflect.Bool {
		return nil
	}

	requiresInit := false
	boolState := core.DerivedState[bool](ctx.State(), ctx.Field().Name).Init(func() bool {
		src := ctx.State().Get()
		v := reflect.ValueOf(src).FieldByName(ctx.Field().Name).Bool()
		if val := ctx.Value(); val != "" && v == false {
			p, err := strconv.ParseBool(val)
			if err == nil {
				requiresInit = true
				return p
			}
		}

		return v
	})

	boolState.Observe(func(newValue bool) {
		ctx.SetValue(newValue)

	})

	if requiresInit {
		boolState.Notify()
	}

	return ui.CheckboxField(ctx.Label(), boolState.Get()).
		Disabled(ctx.Disabled()).
		ID(ctx.ID()).
		ErrorText(ctx.ErrorText()).
		InputValue(boolState).
		SupportingText(ctx.SupportingText()).
		Frame(ui.Frame{}.FullWidth())
	
}
