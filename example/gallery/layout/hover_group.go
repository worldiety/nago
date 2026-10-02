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
	app.Register("hover-group", func(wnd core.Window) core.View {
		return HoverGroup(
			VStack(Text("Hover me")).Frame(Frame{}.Size(L200, L120)),
			VStack(
				Text("More details"),
				PrimaryButton(nil).Title("Open"),
			).Gap(L8).Frame(Frame{}.Size(L200, L120)),
		).BackgroundColor(M2).
			Border(Border{}.Radius(L16)).
			Frame(Frame{}.Size(L200, L120))
	})
}
