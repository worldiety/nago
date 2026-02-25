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
	"go.wdy.de/nago/presentation/ui/timeframe"
)

func RenderXTimeTimeFrame(ctx FieldContext) core.View {
	if !ctx.IsType(reflect.TypeFor[xtime.TimeFrame]()) {
		return nil
	}

	dateState := core.DerivedState[xtime.TimeFrame](ctx.State(), ctx.Field().Name).Init(func() xtime.TimeFrame {
		src := ctx.State().Get()
		v := reflect.ValueOf(src).FieldByName(ctx.Field().Name).Interface()

		d := v.(xtime.TimeFrame)

		return d
	})

	dateState.Observe(func(newValue xtime.TimeFrame) {
		ctx.SetValue(newValue)
	})

	return timeframe.Picker(ctx.Label(), dateState).
		Disabled(ctx.Disabled()).
		SupportingText(ctx.SupportingText()).
		ErrorText(ctx.ErrorText()).
		Frame(ui.Frame{}.FullWidth())

}
