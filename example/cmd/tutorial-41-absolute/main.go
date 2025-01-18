package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "8fb8724f-e604-444c-9671-58d07dd76164"))

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {

			return ui.VStack(
				createLayout(makeSchedule(), 30)...,
			).
				// set the parent to offset/relative to become a parent for the contained absolute positions
				Position(ui.Position{Type: ui.PositionOffset}).
				BackgroundColor(ui.ColorCardTop).
				Frame(ui.Frame{}.Size("60rem", "30rem"))

		}))

	}).Run()
}

func createLayout(schedule []Class, maxRemHeight float64) []core.View {
	var res []core.View
	for i, class := range schedule {
		top := fmt.Sprintf("%.2frem", class.Start*maxRemHeight)
		height := fmt.Sprintf("%.2frem", (class.End-class.Start)*maxRemHeight)
		res = append(res, ui.HStack(
			ui.Text(class.Title).Color(ui.ColorBlack),
		).
			Position(ui.Position{
				Type: ui.PositionAbsolute,
				// shift them to the right for better overlap perception
				Left: ui.Length(fmt.Sprintf("%.2frem", 8*float64(i))),
				// start at our custom top offset
				Top: ui.Length(top),
			}).
			BackgroundColor("#ffffffaa").
			Border(ui.Border{}.Width(ui.L1).Radius(ui.L8).Color("#ffffff")).
			// better use explicit width and height, due to offset semantics of right and bottom
			Frame(ui.Frame{}.Size(ui.L320, ui.Length(height))),
		)
	}

	return res
}

type Class struct {
	Title string
	Start float64
	End   float64
}

func makeSchedule() []Class {
	return []Class{
		{"Verteidigung gegen die dunklen Künste", 0.083, 0.167},
		{"Zaubertränke", 0.183, 0.237},
		{"Verwandlung", 0.213, 0.467},
		{"Kräuterkunde", 0.383, 0.467},
		{"Geschichte der Zauberei", 0.483, 0.567},
		{"Astronomie", 0.283, 0.667},
		{"Pflege magischer Geschöpfe", 0.683, 0.767},
		{"Flugunterricht", 0.583, 0.867},
		{"Arithmantik", 0.843, 0.967},
	}
}
