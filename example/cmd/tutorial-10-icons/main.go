package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/maps"
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	flowbiteSolid "go.wdy.de/nago/presentation/icons/flowbite/solid"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	materialFilled "go.wdy.de/nago/presentation/icons/material/filled"
	materialOutlined "go.wdy.de/nago/presentation/icons/material/outlined"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

type IconSet struct {
	Icons []Icon
}

type Icon struct {
	Name            string
	UsesStrokeColor bool
	SVG             ora.SVG
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {

			return ui.VStack(
				ui.Text("Hero Outline"),
				Preview(makeIconSet(true, heroOutline.All)),

				ui.Text("Hero Solid"),
				Preview(makeIconSet(false, heroSolid.All)),

				ui.Text("Flowbite Outline"),
				Preview(makeIconSet(true, flowbiteOutline.All)),

				ui.Text("Flowbite Solid"),
				Preview(makeIconSet(false, flowbiteSolid.All)),

				ui.Text("Material Filled"),
				Preview(makeIconSet(false, materialFilled.All)),

				ui.Text("Material Outlined"),
				Preview(makeIconSet(false, materialOutlined.All)),
			).Frame(ora.Frame{}.FullWidth())
		})
	}).Run()
}

func Card(ico Icon) core.View {
	return ui.VStack(
		ui.Text(ico.Name),
		ui.With(ui.Image().Embed(ico.SVG), func(image ui.TImage) ui.TImage {
			if ico.UsesStrokeColor {
				return image.StrokeColor("#ff0000")
			}

			return image.FillColor("#ff0000")
		}).Padding(ora.Padding{}.All(ora.L4)).
			Border(ora.Border{}.Circle().Color("#00ff00").Width(ora.L4)).
			Frame(ora.Frame{}.Size(ora.L44, ora.L44)),
	)
}

func Preview(set IconSet) core.View {
	return ui.Grid(slices.Collect(func(yield func(cell ui.TGridCell) bool) {
		for _, icon := range set.Icons {
			yield(ui.GridCell(Card(icon)))
		}
	})...).Gap(ora.L8).Columns(6)
}

func makeIconSet(stroke bool, icons map[string]ora.SVG) IconSet {
	res := IconSet{}
	for _, key := range maps.SortedKeys(icons) {
		res.Icons = append(res.Icons, Icon{
			Name:            key,
			SVG:             icons[key],
			UsesStrokeColor: stroke,
		})
	}

	return res
}
