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
	icons "go.wdy.de/nago/presentation/icons/hero/outline"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/hero"
)

func init() {
	app.Register("hero", func(wnd core.Window) core.View {
		return hero.Hero("Ship your next app in days").
			Subtitle("Nago renders your Go code as a modern web app. No JavaScript, no REST API, no build pipeline.").
			Actions(
				PrimaryButton(func() {}).Title("Get started").PostIcon(icons.ArrowRight),
				SecondaryButton(func() {}).Title("Read the docs"),
			).
			SideSVG(icons.RocketLaunch).
			Alignment(Leading).
			Frame(Frame{Width: L1200, MaxWidth: Full})
	})
}
