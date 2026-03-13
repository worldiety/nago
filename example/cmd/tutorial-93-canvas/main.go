// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"context"
	_ "embed"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/canvas"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed morty_vanilla.png
var morty application.StaticBytes

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_93")
		cfg.Serve(vuejs.Dist())

		mortyUrl := cfg.Resource(morty)

		cfg.RootView(".", func(wnd core.Window) core.View {
			const myCanvas = "."
			const fps = 30

			canvasCtx := canvas.Context2D(wnd, myCanvas)
			var evt core.InputEvent
			wnd.AddInputListener(myCanvas, func(e core.InputEvent) {
				evt = e
				canvasCtx.LoadImage(1, mortyUrl)
				redraw(canvasCtx, evt)
			})

			core.OnFrame(wnd, myCanvas, fps, func(ctx context.Context) {
				//redraw(canvasCtx, evt)
			})

			return ui.VStack(
				canvas.Canvas(myCanvas).Frame(ui.Frame{Width: ui.L400, Height: ui.L320}),
			).Frame(ui.Frame{}.MatchScreen())
		})

	}).Run()
}

func redraw(ctx canvas.TContext2D, evt core.InputEvent) {

	ctx.NewList(1)
	ctx.ClearRect(0, 0, 1000, 1000)
	ctx.FillColor("#ff0000")
	ctx.FillRect(evt.X, evt.Y, 100, 100)
	ctx.DrawImage(1, evt.X, evt.Y)
	ctx.EndList()

	ctx.CallList(1)
}
