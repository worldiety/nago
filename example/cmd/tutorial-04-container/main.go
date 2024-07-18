package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

const (
	red   = "#ff0000"
	green = "#00ff00"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())
		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.HStack(
				withTitle("box", box()),
				withTitle("vstack", vstack()),
				withTitle("hstack", hstack()),
			).Alignment(ora.Top).Frame(ora.Frame{}.FullWidth())

		})

	}).Run()
}

func box() core.View {
	return ui.Box(ui.BoxLayout{
		Top:            ui.Text("top").BackgroundColor(red),
		Center:         ui.Text("center").BackgroundColor(red),
		Bottom:         ui.Text("bottom").BackgroundColor(red),
		Leading:        ui.Text("leading").BackgroundColor(red),
		Trailing:       ui.Text("trailing").BackgroundColor(red),
		TopLeading:     ui.Text("top-leading").BackgroundColor(red),
		TopTrailing:    ui.Text("top-trailing").BackgroundColor(red),
		BottomLeading:  ui.Text("bottom-leading").BackgroundColor(red),
		BottomTrailing: ui.Text("bottom-trailing").BackgroundColor(red),
	}).BackgroundColor(green).Frame(ora.Frame{}.Size(ora.L320, ora.L320))
}

func vstack() core.View {
	return ui.VStack(
		slices.Collect[core.View](func(yield func(view core.View) bool) {
			for _, alignment := range ora.Alignments() {
				yield(withTitle(fmt.Sprintf("vstack %s", alignment.String()),
					ui.VStack(someViews()...).
						Alignment(alignment).
						BackgroundColor(green).
						Frame(ora.Frame{}.Size(ora.L200, ora.L200)),
				))
			}
		})...,
	)
}

func hstack() core.View {
	return ui.VStack(
		slices.Collect[core.View](func(yield func(view core.View) bool) {
			for _, alignment := range ora.Alignments() {
				yield(withTitle(fmt.Sprintf("hstack %s", alignment.String()),
					ui.HStack(someViews()...).
						Alignment(alignment).
						BackgroundColor(green).
						Frame(ora.Frame{}.Size(ora.L200, ora.L200)),
				))
			}
		})...,
	)
}

func withTitle(title string, view core.View) core.View {
	return ui.VStack(
		ui.Text(title).Font(ora.Title),
		view,
	)
}

func someViews() []core.View {
	return []core.View{
		ui.Text("1").BackgroundColor(red).Frame(ora.Frame{}.Size(ora.L16, ora.L16)),
		ui.Text("2").BackgroundColor(red).Frame(ora.Frame{}.Size(ora.L20, ora.L20)),
		ui.Text("3").BackgroundColor(red).Frame(ora.Frame{}.Size(ora.L40, ora.L40)),
	}
}
