// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/xmaps"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	flowbiteSolid "go.wdy.de/nago/presentation/icons/flowbite/solid"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	materialFilled "go.wdy.de/nago/presentation/icons/material/filled"
	materialOutlined "go.wdy.de/nago/presentation/icons/material/outlined"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"slices"
)

type IconSet struct {
	Icons []Icon
}

type Icon struct {
	Name            string
	UsesStrokeColor bool
	SVG             core.SVG
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {

			return VStack(
				Text("Hero Outline"),
				Preview(makeIconSet(true, heroOutline.All)),

				Text("Hero Solid"),
				Preview(makeIconSet(false, heroSolid.All)),

				Text("Flowbite Outline"),
				Preview(makeIconSet(true, flowbiteOutline.All)),

				Text("Flowbite Solid"),
				Preview(makeIconSet(false, flowbiteSolid.All)),

				Text("Material Filled"),
				Preview(makeIconSet(false, materialFilled.All)),

				Text("Material Outlined"),
				Preview(makeIconSet(false, materialOutlined.All)),
			).Frame(Frame{}.FullWidth())
		})
	}).Run()
}

func Card(ico Icon) core.View {
	return VStack(
		Text(ico.Name),
		With(Image().Embed(ico.SVG), func(image TImage) TImage {
			if ico.UsesStrokeColor {
				return image.StrokeColor("#ff0000")
			}

			return image.FillColor("#ff0000")
		}).Padding(Padding{}.All(L4)).
			Border(Border{}.Circle().Color("#00ff00").Width(L4)).
			Frame(Frame{}.Size(L44, L44)),
	)
}

func Preview(set IconSet) core.View {
	return Grid(slices.Collect(func(yield func(cell TGridCell) bool) {
		for _, icon := range set.Icons {
			yield(GridCell(Card(icon)))
		}
	})...).Gap(L8).Columns(6)
}

func makeIconSet(stroke bool, icons map[string]core.SVG) IconSet {
	res := IconSet{}
	for _, key := range xmaps.SortedKeys(icons) {
		res.Icons = append(res.Icons, Icon{
			Name:            key,
			SVG:             icons[key],
			UsesStrokeColor: stroke,
		})
	}

	return res
}
