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
	app.Register("radio-button", func(wnd core.Window) core.View {
		group := AutoRadioStateGroup(wnd, "colors", 3).InitIndex(0)

		return HStack(
			Each2(group.All(), func(idx int, checked *core.State[bool]) core.View {
				return RadioButton(checked.Get()).InputChecked(checked)
			})...,
		).Gap(L16)
	})
}
