// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/document"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_68")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		option.Must(option.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			edit := core.AutoState[bool](wnd)
			text := core.AutoState[string](wnd).Init(func() string {
				return "finally a text to edit"
			})

			d := time.Date(2025, 7, 9, 12, 42, 0, 0, time.Local)

			addComment := core.AutoState[string](wnd).Observe(func(newValue string) {
				alert.ShowBannerMessage(wnd, alert.Message{Title: "Kommentar hinzugefügt", Message: newValue, Intent: alert.IntentOk})
			})

			selectedA := core.AutoState[bool](wnd)
			selectedB := core.AutoState[bool](wnd)

			return ui.VStack(
				ui.H1("Document example"),
				document.Page(
					document.WithCommentSelection(selectedA, ui.Text("hello word")),
					document.Editable(
						func() core.View {
							return document.WithCommentSelection(selectedB, ui.Text(text.Get()))
						},

						func() core.View {
							return ui.TextField("text", text.Get()).InputValue(text)
						},
					).Style(document.TopTrailing).
						InputValue(edit),
				).Size(document.DinA4).
					Comment(
						document.LogEntry("hello log entry", "Torben", d),
						document.LogEntry("another log entry with a lot more text to tell and more and more stories about stuff and things which don't matter at all.", "Olaf", d),
						document.Thread(wnd, document.Message{User: wnd.Subject().ID(), Message: "Toller Einzelkommentar", Time: d}).InputSelectedValue(selectedA),
						document.Thread(wnd,
							document.Message{User: wnd.Subject().ID(), Message: "Wie findest du das?", Time: d},
							document.Message{User: wnd.Subject().ID(), Message: "so mäßig", Time: d},
							document.Message{User: wnd.Subject().ID(), Message: "wieso?", Time: d},
							document.Message{User: wnd.Subject().ID(), Message: "sieht aus wie word", Time: d},
						).InputValue(addComment).InputSelectedValue(selectedB).Resolve(func() {
							alert.ShowBannerMessage(wnd, alert.Message{Title: "Erledigt", Message: "Gut gemacht", Intent: alert.IntentOk})
						}),
					),
			).FullWidth().Alignment(ui.Leading)

		})
	}).
		Run()
}
