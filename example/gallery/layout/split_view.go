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
	app.Register("split-view", func(wnd core.Window) core.View {
		ratio := core.AutoState[float64](wnd).Init(func() float64 {
			return 0.3
		})

		return SplitView(
			VStack(Text("Content A")).BackgroundColor("#C9E7F8").Frame(Frame{}.FullWidth().FullHeight()),
			VStack(Text("Content B")).BackgroundColor("#FDE2C4").Frame(Frame{}.FullWidth().FullHeight()),
		).InputValue(ratio).
			MinRatio(0.2).
			MaxRatio(0.8).
			Frame(Frame{}.Size(L560, L200))
	})
}
