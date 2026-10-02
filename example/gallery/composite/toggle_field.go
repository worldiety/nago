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
	app.Register("toggle-field", func(wnd core.Window) core.View {
		newsletter := core.AutoState[bool](wnd).Init(func() bool { return true })
		tracking := core.AutoState[bool](wnd)

		return VStack(
			ToggleField("Newsletter", newsletter.Get()).
				InputValue(newsletter).
				SupportingText("Receive product news once a month"),
			ToggleField("Usage statistics", tracking.Get()).
				InputValue(tracking).
				ErrorText("Required by your organization"),
		).Alignment(Leading).Gap(L16)
	})
}
