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
	"go.wdy.de/nago/presentation/ui/alert"
)

func init() {
	app.Register("banner-messages", func(wnd core.Window) core.View {
		return VStack(
			PrimaryButton(func() {
				alert.ShowBannerMessage(wnd, alert.Message{
					Title:   "Saved",
					Message: "Your changes have been saved.",
					Intent:  alert.IntentSuccess,
				})
			}).Title("Save"),

			// renders all pending messages as an overlay
			alert.BannerMessages(wnd),
		)
	})
}
