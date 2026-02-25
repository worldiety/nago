// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"reflect"

	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func RenderXTimeDate(ctx FieldContext) core.View {
	if !ctx.IsType(reflect.TypeFor[xtime.Date]()) {
		return nil
	}

	dateState := core.DerivedState[xtime.Date](ctx.State(), ctx.Field().Name).Init(func() xtime.Date {
		src := ctx.State().Get()
		v := reflect.ValueOf(src).FieldByName(ctx.Field().Name).Interface()

		d := v.(xtime.Date)

		return d
	})

	dateState.Observe(func(newValue xtime.Date) {
		ctx.SetValue(newValue)
	})

	return ui.SingleDatePicker(ctx.Label(), dateState.Get(), dateState).
		Disabled(ctx.Disabled()).
		SupportingText(ctx.SupportingText()).
		ErrorText(ctx.ErrorText()).
		Frame(ui.Frame{}.FullWidth())

}
