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
	"go.wdy.de/nago/presentation/ui/avatar"
)

func init() {
	app.Register("avatar", func(wnd core.Window) core.View {
		return HStack(
			avatar.Text("Ada Lovelace"),
			avatar.Text("Grace Hopper").Size(L64),
			avatar.Text("Alan Turing").Size(L80).Style(avatar.Rounded),
			avatar.Text("Linus").Size(L96).Action(func() {
				// open the profile
			}),
		).Gap(L24)
	})
}
