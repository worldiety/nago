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
	app.Register("frame", func(wnd core.Window) core.View {
		return VStack(
			HStack(
				Text("Size").BackgroundColor("#C9E7F8").Frame(Frame{}.Size(L120, L48)),
				Text("Grow()").BackgroundColor("#FDE2C4").Frame(Frame{Height: L48}.Grow()),
				Text("Size").BackgroundColor("#C9E7F8").Frame(Frame{}.Size(L120, L48)),
			).Gap(L8).FullWidth(),
			Text("FullWidth()").BackgroundColor("#D8F0D2").Frame(Frame{Height: L48}.FullWidth()),
			Text("MinWidth: L200").BackgroundColor("#E5D8F6").Frame(Frame{MinWidth: L200, Height: L48}),
		).Gap(L8).
			Alignment(Leading).
			Frame(Frame{Width: L560})
	})
}
