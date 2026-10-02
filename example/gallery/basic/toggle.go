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
	app.Register("toggle", func(wnd core.Window) core.View {
		enabled := core.AutoState[bool](wnd).Init(func() bool { return true })

		return HStack(
			Toggle(enabled.Get()).InputChecked(enabled),
			Toggle(false),
			Toggle(true).Disabled(true),
		).Gap(L16)
	})
}
