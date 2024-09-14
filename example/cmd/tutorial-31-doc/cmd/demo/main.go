// # Terminplanung-Service
//
// Die Domäne der Terminplanung umfasst die Methoden, Werkzeuge und Prozesse, die zur Verwaltung und Organisation von
// Terminen und Zeitplänen verwendet werden. Sie ist in vielen Bereichen von
// entscheidender Bedeutung, darunter im Gesundheitswesen, in der Unternehmensverwaltung,
// im Bildungswesen und in der persönlichen Organisation.
//
// ## Subkapitel
//
// Weiterer Text und es folgt eine Aufzählung:
//   - eins
//   - zwei
//   - drei
//
// Davon abzugrenzen ist ein pre-formatted block:
//
//	err := xreflect.Import(xreflect.ModName()+"/example/cmd/tutorial-31-doc", tutorial_31_doc.Src)
//	if err != nil {
//	  panic(err)
//	}
//
// Tolles Showbild ist in ![alt text](https://cdn.prod.website-files.com/65c9e1a09853d67c47d4320d/66759ca28472aeb9389f94b5_worldiety-team.jpg) zu sehen.
// Weitere Information im Bild oberhalb.
package main

import (
	"go.wdy.de/nago/application"
	tutorial_31_doc "go.wdy.de/nago/example/cmd/tutorial-31-doc"
	"go.wdy.de/nago/glossary"
	"go.wdy.de/nago/glossary/docm/oraui"
	"go.wdy.de/nago/pkg/xreflect"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"

	"go.wdy.de/nago/web/vuejs"
)

func main() {
	err := xreflect.Import(xreflect.ModName()+"/example/cmd/tutorial-31-doc", tutorial_31_doc.Src)
	if err != nil {
		panic(err)
	}

	// we use the applications package to bootstrap our configuration
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		doc := glossary.Auto()

		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.VStack(ui.HStack(oraui.Render(doc)).Frame(ui.Frame{MaxWidth: "900dp"})).FullWidth()
		})
	}).
		Run()
}
