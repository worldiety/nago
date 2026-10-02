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
	app.Register("dnd", func(wnd core.Window) core.View {
		dropped := core.AutoState[string](wnd).Observe(func(id string) {
			// the area with this id was dropped onto the zone
		})

		item := func(id, label string) core.View {
			return DnDArea(
				VStack(Text(label)).BackgroundColor("#C9E7F8").Frame(Frame{}.Size(L80, L80)),
			).ID(id).CanDrag(true)
		}

		return HStack(
			VStack(item("a", "A"), item("b", "B")).Gap(L16),
			DnDArea(
				VStack(Text("Drop A here")).BackgroundColor("#FDE2C4").Frame(Frame{}.Size(L200, L200)),
			).ID("zone").
				CanDrop(true).
				Droppable("a").
				InputValue(dropped),
		).Gap(L32)
	})
}
