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
	app.Register("position", func(wnd core.Window) core.View {
		badge := HStack(Text("New").Color(ColorWhite)).
			Position(Position{Type: PositionAbsolute, Top: L8, Right: L8}).
			BackgroundColor("#FA2C7F").
			Padding(Padding{}.Horizontal(L8)).
			Border(Border{}.Radius(L8))

		return VStack(
			Text("The badge is placed absolutely within this card."),
			badge,
		).
			// the parent must use PositionOffset to become the anchor of absolute children
			Position(Position{Type: PositionOffset}).
			BackgroundColor(ColorCardBody).
			Padding(Padding{}.All(L16)).
			Frame(Frame{}.Size(L320, L160))
	})
}
