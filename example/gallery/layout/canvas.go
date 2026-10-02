// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"math"

	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/canvas"
)

func init() {
	app.Register("canvas", func(wnd core.Window) core.View {
		const id = "drawing"

		ctx := canvas.Context2D(wnd, id)
		// draw a scene and a circle wherever the user clicks
		wnd.AddInputListener(id, func(evt core.InputEvent) {
			ctx.Clear()
			ctx.FillColor("#C9E7F8").FillRect(0, 0, 400, 200)
			ctx.Font("20px sans-serif").FillColor("#333333").FillText("Click to draw", 20, 40, 300)
			ctx.BeginPath().Arc(evt.X, evt.Y, 30, 0, 2*math.Pi, false).FillColor("#F7A823").Fill()
		}, core.InputEventPointerDown)

		return canvas.Canvas(id).Frame(Frame{Width: L400, Height: L200})
	})
}
