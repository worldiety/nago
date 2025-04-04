// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"slices"
)

// Card creates a card view based on the field bindings. The given value is mapped automatically, based on the binding.
// A Card is usually readonly.
func Card[T any](bnd *Binding[T], value *core.State[T]) ui.DecoredView {
	return ui.VStack(
		slices.Collect(func(yield func(view core.View) bool) {
			for _, field := range bnd.fields {
				if field.RenderCardElement != nil {
					yield(field.RenderCardElement(field, value).Frame(ui.Frame{}.FullWidth()))
				}
			}

		})...,
	).Gap(ui.L16).
		BackgroundColor(ui.ColorCardBody).
		Padding(ui.Padding{}.All(ui.L20)).
		Border(ui.Border{}.Radius(ui.L16))
}
