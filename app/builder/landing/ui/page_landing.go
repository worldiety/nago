// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uilanding

import (
	_ "embed"

	uienv "go.wdy.de/nago/app/builder/environment/ui"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/hero"
	"go.wdy.de/nago/presentation/ui/markdown"
)

//go:embed screenshot.png
var TeaserImg application.StaticBytes

var (
	StrSubtitle = core.DefaultStr("nbuilder.landing.hero.subtitle", "Create Scaffolds and entire applications without coding.", "Erstelle Projektrümpfe oder sogar ganze Anwendungen ohne zu programmieren.")
	StrStartNow = core.DefaultStr("nbuilder.landing.hero.cta", "Start now", "Jetzt loslegen")
	StrWorkflow = core.DefaultStr("nbuilder.landing.workflow",
		`# Workflow

1. Create an environment where you can build your apps and collaborate with others.
An environment can be a project or a workspace for your entire company.
2. Define your data models and model persistent data using repositories to quickly model CRUD scenarios.
3. Define use cases so that you can develop them further later using full code.
4. Build on ready-made templates for the interface design and try out the designer. Then switch to the generated code to design dynamic interfaces as you wish using full code.
5. Use AI to generate suggestions and create an initial draft.
`,
		`
# Workflow

1. Erstelle eine Umgebung in der du deine Apps anlegst und wo du mit mehreren zusammenarbeiten kannst.
Eine Umgebung kann ein Projekt sein oder ein Arbeitsbereich für deine ganze Firma.
2. Definiere deine Datenmodelle und modelliere persistente Daten mit Hilfe von Repositories, um schnell CRUD-Szenarien zu modellieren.
3. Definiere Anwendungsfälle, um diese später mittels Full-Code mit professionell weiter zu entwickeln.
4. Baue auf fertigen Templates für das Oberflächendesign auf und probiere den Designer. Wechsle danach in den generierten Code, um dynamische Oberflächen nach Belieben mittel Full-Code zu gestalten.
5. Verwende KI um dir Vorschläge erstellen zu lassen und einen ersten Aufschlag zu generieren.
`,
	)
)

func PageLanding(wnd core.Window, teaser core.URI) core.View {
	return ui.VStack(

		hero.Hero("nago builder").
			Alignment(ui.BottomLeading).
			Subtitle(StrSubtitle.Get(wnd)).
			SideSVG(icons.QrCode).
			BackgroundImage(teaser).
			ForegroundColorAdaptive("#00000066", "#ffffff66").
			Actions(ui.PrimaryButton(func() {
				wnd.Navigation().ForwardTo(uienv.PathEnvironments, nil)
			}).Title(StrStartNow.Get(wnd))),

		ui.Space(ui.L48),

		markdown.RichText(StrWorkflow.Get(wnd)).Frame(ui.Frame{}.Larger()),
	).FullWidth()
}
