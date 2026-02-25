// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"reflect"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/colorpicker"
)

func RenderColor(ctx FieldContext) core.View {
	if !ctx.IsType(reflect.TypeFor[ui.Color]()) {
		return nil
	}
	
	requiresInit := false
	colorState := core.DerivedState[ui.Color](ctx.State(), ctx.Field().Name).Init(func() ui.Color {
		src := ctx.State().Get()
		str := reflect.ValueOf(src).FieldByName(ctx.Field().Name).String()
		if val := ctx.Value(); val != "" && str == "" {
			requiresInit = true
			return ui.Color(val)
		}

		return ui.Color(str)
	})

	colorState.Observe(func(newValue ui.Color) {
		ctx.SetValue(newValue)
	})

	if requiresInit {
		colorState.Notify()
	}

	return colorpicker.PalettePicker(ctx.Label(), colorpicker.DefaultPalette).
		State(colorState).
		SupportingText(ctx.supportingText).
		Disabled(ctx.Disabled()).
		Value(colorState.Get()).
		ErrorText(ctx.errorText).
		Frame(ui.Frame{}.FullWidth())
}
