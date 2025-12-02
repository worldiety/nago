// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uilanding

import (
	_ "embed"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/hero"
)

//go:embed screenshot.png
var TeaserImg application.StaticBytes

func PageLanding(wnd core.Window, teaser core.URI) core.View {
	return ui.VStack(

		hero.Hero("nago builder").
			Alignment(ui.BottomLeading).
			Subtitle("Scaffolds und Apps erstellen afds kfn dsakjf ;sadfkj laskdf ;dkjf ladskj f;lsadjf; dsf kdf ").
			SideSVG(icons.QrCode).
			BackgroundImage(teaser).
			ForegroundColor("#000000aa").
			Actions(ui.PrimaryButton(nil).Title("jetzt loslegen")),
		//Teaser(ui.Image().URI(teaser)).
		//Frame(ui.Frame{Height: ui.L560}),
	).Frame(ui.Frame{}.Larger())
}
