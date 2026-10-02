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
	app.Register("transformation", func(wnd core.Window) core.View {
		return HStack(
			VStack(Text("none")).
				BackgroundColor("#C9E7F8").
				Frame(Frame{}.Size(L120, L80)),
			VStack(Text("RotateZ: 15")).
				Transformation(Transformation{RotateZ: 15}).
				BackgroundColor("#FDE2C4").
				Frame(Frame{}.Size(L120, L80)),
			VStack(Text("ScaleX: 0.7")).
				Transformation(Transformation{ScaleX: 0.7, ScaleY: 1}).
				BackgroundColor("#D8F0D2").
				Frame(Frame{}.Size(L120, L80)),
			VStack(Text("TranslateY: L16")).
				Transformation(Transformation{TranslateY: L16}).
				BackgroundColor("#E5D8F6").
				Frame(Frame{}.Size(L120, L80)),
		).Gap(L32).Padding(Padding{}.All(L16))
	})
}
