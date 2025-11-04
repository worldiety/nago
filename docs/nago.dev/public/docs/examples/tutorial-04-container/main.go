// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"slices"
)

const (
	red   = "#ff0000"
	green = "#00ff00"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())
		cfg.RootView(".", func(wnd core.Window) core.View {
			return HStack(
				withTitle("box", box()),
				withTitle("vstack", vstack()),
				withTitle("hstack", hstack()),
			).Alignment(Top).Frame(Frame{}.FullWidth())

		})

	}).Run()
}

func box() core.View {
	return Box(BoxLayout{
		Top:            Text("top").BackgroundColor(red),
		Center:         Text("center").BackgroundColor(red),
		Bottom:         Text("bottom").BackgroundColor(red),
		Leading:        Text("leading").BackgroundColor(red),
		Trailing:       Text("trailing").BackgroundColor(red),
		TopLeading:     Text("top-leading").BackgroundColor(red),
		TopTrailing:    Text("top-trailing").BackgroundColor(red),
		BottomLeading:  Text("bottom-leading").BackgroundColor(red),
		BottomTrailing: Text("bottom-trailing").BackgroundColor(red),
	}).BackgroundColor(green).Frame(Frame{}.Size(L320, L320))
}

func vstack() core.View {
	return VStack(
		slices.Collect[core.View](func(yield func(view core.View) bool) {
			for _, alignment := range Alignments() {
				yield(withTitle(fmt.Sprintf("vstack %s", alignment.String()),
					VStack(someViews()...).
						Alignment(alignment).
						BackgroundColor(green).
						Frame(Frame{}.Size(L200, L200)),
				))
			}
		})...,
	)
}

func hstack() core.View {
	return VStack(
		slices.Collect[core.View](func(yield func(view core.View) bool) {
			for _, alignment := range Alignments() {
				yield(withTitle(fmt.Sprintf("hstack %s", alignment.String()),
					HStack(someViews()...).
						Alignment(alignment).
						BackgroundColor(green).
						Frame(Frame{}.Size(L200, L200)),
				))
			}
		})...,
	)
}

func withTitle(title string, view core.View) core.View {
	return VStack(
		Text(title).Font(Title),
		view,
	)
}

func someViews() []core.View {
	return []core.View{
		Text("1").BackgroundColor(red).Frame(Frame{}.Size(L16, L16)),
		Text("2").BackgroundColor(red).Frame(Frame{}.Size(L20, L20)),
		Text("3").BackgroundColor(red).Frame(Frame{}.Size(L40, L40)),
	}
}
