package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/list"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.VStack(
				list.List(
					list.Entry().
						Leading(ui.MailTo(wnd, "test@test.de", "test@test.de")).
						Headline("Bilbo").
						SupportingText("Ein Beutlin.").
						Trailing(ui.ImageIcon(heroSolid.ArrowRight)),
					list.Entry().
						Leading(ui.ImageIcon(heroSolid.XMark)).
						Headline("Gollumn").
						SupportingText("Ein Hobbit.").
						Trailing(ui.ImageIcon(heroSolid.ArrowRight)),
				).Caption(ui.Text("Alle Teilnehmer")).
					Footer(ui.Text("2 Einträge")).
					Frame(ui.Frame{Width: ui.L400}),
			).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
