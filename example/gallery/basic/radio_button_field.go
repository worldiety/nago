// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
)

func init() {
	app.Register("radio-button-field", func(wnd core.Window) core.View {
		group := AutoRadioStateGroup(wnd, "sizes", 3).InitIndex(1)
		labels := []string{"Small", "Medium", "Large"}

		return VStack(
			Each2(group.All(), func(idx int, checked *core.State[bool]) core.View {
				return RadioButtonField(labels[idx], &group, idx)
			})...,
		).Alignment(Leading).Gap(L8)
	})
}
