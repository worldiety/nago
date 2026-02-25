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
	"strings"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func RenderSliceText(ctx FieldContext) core.View {
	field := ctx.Field()
	if !ctx.IsSlice(reflect.String) || ctx.Source() != nil {
		return nil
	}

	var lines int
	if str, ok := field.Tag.Lookup("lines"); ok {
		lines, _ = strconv.Atoi(str)
	}

	if lines == 0 {
		lines = 5
	}

	requiresInit := false
	strState := core.DerivedState[string](ctx.State(), field.Name).Init(func() string {
		src := ctx.State().Get()
		slice := reflect.ValueOf(src).FieldByName(field.Name)
		tmp := make([]string, 0, slice.Len())
		for i := 0; i < slice.Len(); i++ {
			tmp = append(tmp, slice.Index(i).String())
		}

		str := strings.Join(tmp, "\n")

		if val := field.Tag.Get("value"); val != "" && str == "" {
			requiresInit = true
			return ctx.Window().Bundle().Resolve(val)
		}

		return str
	})

	strState.Observe(func(newValue string) {
		v := strings.Split(newValue, "\n")
		slice := reflect.MakeSlice(field.Type, 0, len(v))
		for _, strVal := range v {
			newValue := reflect.New(field.Type.Elem()).Elem()
			newValue.SetString(strVal)
			slice = reflect.Append(slice, newValue)
		}

		ctx.SetValue(slice.Interface())
	})

	if requiresInit {
		strState.Notify()
	}

	return ui.TextField(ctx.Label(), strState.Get()).
		InputValue(strState).
		ID(ctx.ID()).
		SupportingText(ctx.SupportingText()).
		Lines(lines).
		ErrorText(ctx.ErrorText()).
		Disabled(ctx.Disabled()).
		FullWidth()
}
