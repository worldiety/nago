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
	app.Register("spacer", func(wnd core.Window) core.View {
		return HStack(
			Text("Leading").BackgroundColor("#C9E7F8").Padding(Padding{}.All(L16)),
			Spacer(),
			Text("Trailing").BackgroundColor("#FDE2C4").Padding(Padding{}.All(L16)),
		).BackgroundColor(ColorCardBody).
			Padding(Padding{}.All(L8)).
			Frame(Frame{Width: L560})
	})
}
