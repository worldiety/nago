// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"fmt"
	"reflect"

	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/avatar"
)

func RenderImage(ctx FieldContext) core.View {
	if !ctx.IsType(reflect.TypeFor[image.ID]()) {
		return nil
	}

	requiresInit := false
	imageState := core.DerivedState[image.ID](ctx.State(), ctx.Field().Name).Init(func() image.ID {
		src := ctx.State().Get()
		str := reflect.ValueOf(src).FieldByName(ctx.Field().Name).String()
		if val := ctx.Value(); val != "" && str == "" {
			requiresInit = true
			return image.ID(val)
		}

		return image.ID(str)
	})

	imageState.Observe(func(newValue image.ID) {
		ctx.SetValue(newValue)
	})

	if requiresInit {
		imageState.Notify()
	}

	var tmp []core.View
	if ctx.Label() != "" {
		tmp = append(tmp, ui.Text(ctx.Label()).TextAlignment(ui.TextAlignStart).FullWidth())
	}

	var picker core.View
	switch ctx.Field().Tag.Get("style") {
	case "avatar":
		picker = AvatarPicker(ctx.Window(), nil, ctx.Field().Name, imageState.Get(), imageState, paraphe(ctx), avatar.Circle).Enabled(!ctx.Disabled())
	case "icon":
		picker = AvatarPicker(ctx.Window(), nil, ctx.Field().Name, imageState.Get(), imageState, paraphe(ctx), avatar.Rounded).Enabled(!ctx.Disabled())
	default:
		picker = SingleImagePicker(ctx.Window(), nil, nil, nil, ctx.Field().Name, imageState.Get(), imageState)
	}

	if len(tmp) == 0 {
		return picker
	}

	return ui.VStack(tmp...).Append(picker).FullWidth()
}

func paraphe(ctx FieldContext) string {
	if stringer, ok := ctx.State().Get().(fmt.Stringer); ok {
		return stringer.String()
	}

	return ctx.Label()
}
