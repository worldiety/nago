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
	app.Register("background", func(wnd core.Window) core.View {
		return HStack(
			VStack(Text("AppendLinearGradient").Color(ColorWhite)).
				Background(Background{}.AppendLinearGradient("#1940BF", "#FA2C7F")).
				Frame(Frame{}.Size(L320, L160)).
				Border(Border{}.Radius(L16)),
			VStack(Text("two layers").Color(ColorWhite)).
				Background(Background{}.
					AppendLinearGradient("#01BA6C", "#1CCDFB").
					AppendLinearGradient("#00000000", "#000000AA"),
				).
				Frame(Frame{}.Size(L320, L160)).
				Border(Border{}.Radius(L16)),
		).Gap(L16).BackgroundColor(ColorCardBody).Padding(Padding{}.All(L16))
	})
}
