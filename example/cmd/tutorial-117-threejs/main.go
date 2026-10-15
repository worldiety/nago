// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Command tutorial-117-threejs shows a three.js scene, which a program of the gostrict subset builds and animates
// in the browser. The program reports clicks to its backend in Go, which answers with a new color, and the
// backend changes the speed of the program at any time. Each scene runs in the browser, so both instances on the
// page animate at the frame rate of the browser without the server.
package main

import (
	"context"
	"embed"
	"fmt"
	"math/rand/v2"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/example/cmd/tutorial-117-threejs/scene"
	"go.wdy.de/nago/example/cmd/tutorial-117-threejs/scene/proto"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/program"
	"go.wdy.de/nago/presentation/ui/threejs"
	"go.wdy.de/nago/web/vuejs"
)

// The sources of the program, which the application translates at the start. The generated gostrict_gen.go is
// part of the backend and ignored by the translation.
//
//go:embed scene
var sources embed.FS

const module = "go.wdy.de/nago/example/cmd/tutorial-117-threejs"

// backend receives the messages of the program in the browser.
type backend struct {
	ch     *program.Channel
	clicks *core.State[int]
}

func (b backend) Clicked(ctx context.Context, m proto.Clicked) error {
	b.clicks.Set(b.clicks.Get() + 1)
	// answer right away with a random color
	return b.ch.Send(scene.ColorMessage(proto.Color{RGB: rand.Uint32N(0xffffff)}))
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_117")
		cfg.Serve(vuejs.Dist())

		bundle := program.MustCompile(module+"/scene", gostrict.Module{Path: module, FS: sources})

		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.H1("three.js with gostrict"),
				knot(wnd, bundle, "knot-a", 12),
				knot(wnd, bundle, "knot-b", 5),
			).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}

func knot(wnd core.Window, bundle *gostrict.Bundle, id string, balls int32) core.View {
	clicks := core.StateOf[int](wnd, id+".clicks")
	ch := program.ChannelOf(wnd, id)
	be := backend{ch: ch, clicks: clicks}

	return ui.VStack(
		threejs.ThreeJS(id, bundle).
			Init("Init", balls).
			Animate("Frame").
			Resize("Resize").
			Antialias(true).
			OnMessage(func(ctx context.Context, m gostrict.Message) error {
				return scene.Dispatch(ctx, be, m)
			}).
			Frame(ui.Frame{Width: ui.L560, Height: ui.L320}),
		ui.HStack(
			ui.Text(fmt.Sprintf("%d clicks", clicks.Get())),
			ui.SecondaryButton(func() {
				_ = ch.Send(scene.SpeedMessage(proto.Speed{Factor: 0.2}))
			}).Title("slow"),
			ui.SecondaryButton(func() {
				_ = ch.Send(scene.SpeedMessage(proto.Speed{Factor: 3}))
			}).Title("fast"),
		).Gap(ui.L8),
	).Gap(ui.L8)
}
