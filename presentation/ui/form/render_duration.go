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
	"time"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/timepicker"
)

func RenderDuration(ctx FieldContext) core.View {
	if !ctx.IsType(reflect.TypeFor[time.Duration]()) {
		return nil
	}

	var displayFormat timepicker.PickerFormat
	switch ctx.Field().Tag.Get("style") {
	case "decomposed":
		displayFormat = timepicker.DecomposedFormat
	case "clock":
		displayFormat = timepicker.ClockFormat
	}

	field := ctx.Field()
	requiresInit := false
	intState := core.DerivedState[time.Duration](ctx.State(), field.Name).Init(func() time.Duration {
		src := ctx.State().Get()
		v := reflect.ValueOf(src).FieldByName(field.Name).Int()
		if val := ctx.Value(); val != "" && v == 0 {
			p, err := strconv.ParseInt(val, 10, 64)
			if err == nil {
				requiresInit = true
				return time.Duration(p)
			}
		}

		return time.Duration(v)
	})

	intState.Observe(func(newValue time.Duration) {
		ctx.SetValue(newValue)
	})

	if requiresInit {
		intState.Notify()
	}

	showDays := true
	if v, ok := field.Tag.Lookup("days"); ok {
		showDays, _ = strconv.ParseBool(v)
	}

	showHours := true
	if v, ok := field.Tag.Lookup("hours"); ok {
		showHours, _ = strconv.ParseBool(v)
	}

	showMinutes := true
	if v, ok := field.Tag.Lookup("minutes"); ok {
		showMinutes, _ = strconv.ParseBool(v)
	}

	showSeconds := true
	if v, ok := field.Tag.Lookup("seconds"); ok {
		showSeconds, _ = strconv.ParseBool(v)
	}

	return timepicker.Picker(ctx.Label(), intState).
		Format(displayFormat).
		Days(showDays).
		Hours(showHours).
		Minutes(showMinutes).
		Seconds(showSeconds).
		Disabled(ctx.Disabled()).
		ErrorText(ctx.ErrorText()).
		SupportingText(ctx.SupportingText()).
		Frame(ui.Frame{}.FullWidth())
}
