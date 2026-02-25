// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"log/slog"
	"reflect"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/picker"
)

func RenderSourceString(ctx FieldContext) core.View {
	field := ctx.Field()
	if ctx.Field().Type.Kind() != reflect.String {
		return nil
	}

	source := ctx.Source()
	if source == nil {
		return nil
	}

	var values []Entity
	for id, err := range source.FindAll(ctx.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		optE, err := source.FindByID(ctx.Subject(), id)
		if err != nil {
			return alert.BannerError(err)
		}

		values = append(values, optE.Unwrap())
	}

	strState := core.DerivedState[[]Entity](ctx.State(), field.Name).Init(func() []Entity {
		src := ctx.State().Get()
		slice := reflect.ValueOf(src).FieldByName(field.Name)
		tmp := make([]Entity, 0, slice.Len())

		for _, id := range slice.Seq2() {
			id := id.String()
			optV, err := source.FindByID(ctx.Subject(), id)
			if err != nil {
				slog.Error("form.Auto failed to resolve source value by id", "id", id, "err", err.Error())
				continue
			}

			if optV.IsNone() {
				slog.Error("form.Auto source has no such entity", "id", id)
				continue
			}

			tmp = append(tmp, optV.Unwrap())
		}

		return tmp
	})

	strState.Observe(func(v []Entity) {
		if len(v) == 0 {
			ctx.SetValue("")
			return
		}

		ctx.SetValue(v[0].ID)
	})

	dlgOpts := getDialogOptions(field)

	return picker.Picker[Entity](ctx.Label(), values, strState).
		Title(ctx.Label()).
		MultiSelect(false). // multi select is another renderer, which only works on non-slice types
		ErrorText(ctx.ErrorText()).
		Disabled(ctx.Disabled()).
		DialogOptions(dlgOpts...).
		SupportingText(ctx.SupportingText()).
		Frame(ui.Frame{}.FullWidth())

}
