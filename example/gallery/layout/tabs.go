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
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/tabs"
)

// the demo is shared by the tabs and page pages
func init() {
	app.Register("tabs", func(wnd core.Window) core.View {
		return tabs.Tabs(
			tabs.Page("Overview", func() core.View {
				return Text("The content of the overview page.")
			}).Icon(icons.Home),
			tabs.Page("Settings", func() core.View {
				return Text("The content of the settings page.")
			}).Icon(icons.Cog6Tooth),
			tabs.Page("Archive", func() core.View {
				return Text("Not available yet.")
			}).Disabled(true),
		).InputValue(core.AutoState[int](wnd)).
			Frame(Frame{Width: L560})
	})
}
