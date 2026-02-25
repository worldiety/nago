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

func RenderText(ctx FieldContext) core.View {
	if ctx.Field().Type.Kind() != reflect.String {
		return nil
	}

	requiresInit := false
	strState := core.DerivedState[string](ctx.State(), ctx.Field().Name).Init(func() string {
		src := ctx.State().Get()
		str := reflect.ValueOf(src).FieldByName(ctx.Field().Name).String()
		if val := ctx.Value(); val != "" && str == "" {
			requiresInit = true
			return ctx.Window().Bundle().Resolve(val)
		}

		return str
	})

	strState.Observe(func(newValue string) {
		ctx.SetValue(newValue)
	})

	if requiresInit {
		strState.Notify()
	}

	var lines int
	if str, ok := ctx.Field().Tag.Lookup("lines"); ok {
		lines, _ = strconv.Atoi(str)
	}

	secretStyle := ctx.Field().Tag.Get("style") == "secret"
	if secretStyle {
		return ui.PasswordField(ctx.Label(), strState.Get()).
			InputValue(strState).
			ID(ctx.ID()).
			SupportingText(ctx.SupportingText()).
			ErrorText(ctx.ErrorText()).
			Disabled(ctx.Disabled()).
			AutoComplete(false).
			FullWidth()
	}

	return ui.TextField(ctx.Label(), strState.Get()).
		InputValue(strState).
		ID(ctx.ID()).
		SupportingText(ctx.SupportingText()).
		ErrorText(ctx.ErrorText()).
		Lines(lines).
		Disabled(ctx.Disabled()).
		FullWidth()
}
