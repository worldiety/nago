// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// This tutorial sends push notifications to the browsers of users. Start it and sign in as the bootstrap admin:
//
//	go run ./example/cmd/tutorial-122-web-push
//
// The start page does not use the notifications, thus nago asks the browser for nothing there. The page
// "notifications" shows [uipush.Card], whose button lets the browser ask for the permission. Only after that click,
// nago registers its service worker and subscribes the browser. The account view offers the same as a switch,
// because of [cfgpush.Options.AccountSection].
//
// Web push requires a secure context, i.e. https or localhost. Safari on iPhone and iPad supports it only for an
// installed web app: tap "Share" and "Add to Home Screen" and open the app from the home screen. Set a manifest
// icon by [application.Configurator.SetManifestIcon] for the home screen and the notifications.
package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/push"
	cfgpush "go.wdy.de/nago/application/push/cfg"
	uipush "go.wdy.de/nago/application/push/ui"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_122")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())
		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		pushes := std.Must(cfgpush.Enable(cfg, cfgpush.Options{
			// also offer a switch in the account view, besides the card of the page below
			AccountSection: true,
		}))

		// this page does not use the notifications, so the browser is never asked here
		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.H1("Tutorial 122: Web Push"),
				ui.Text("Diese Seite nutzt keine Benachrichtigungen. Der Browser fragt hier nach nichts."),
				ui.PrimaryButton(func() {
					wnd.Navigation().ForwardTo("notifications", nil)
				}).Title("Zu den Benachrichtigungen"),
			).Alignment(ui.Leading).Gap(ui.L16).Frame(ui.Frame{}.Larger())
		})

		cfg.RootViewWithDecoration("notifications", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.H1("Benachrichtigungen"),

				// the only place, where the browser asks for the permission: the button of the card
				uipush.Card(wnd),

				ui.Text("Schließe nach dem Klick den Tab oder wechsle die App: Die Benachrichtigung kommt trotzdem. Ein Klick darauf öffnet den Auftrag."),
				ui.HStack(
					ui.SecondaryButton(func() {
						sendLater(wnd, pushes.UseCases)
					}).Title("In 10 Sekunden an mich senden"),

					// Send requires the permission push.PermSend and reaches users, groups and roles, e.g.
					// push.ToGroup("sales") or push.ToRole("dispatcher"). A background job sends as user.SU().
					ui.SecondaryButton(func() {
						report, err := pushes.UseCases.Send(wnd.Subject(), push.ToUser(wnd.Subject().ID()), order("42"))
						if err != nil {
							alert.ShowBannerError(wnd, err)
							return
						}

						alert.ShowBannerMessage(wnd, alert.Message{Title: "Gesendet", Message: describe(report), Intent: alert.IntentOk})
					}).Title("Jetzt per Send an mich senden"),
				).Gap(ui.L8),
			).Alignment(ui.Leading).Gap(ui.L16).Frame(ui.Frame{}.Larger())
		})

		// a click on a notification opens this route with its values
		cfg.RootViewWithDecoration("order", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.H1("Auftrag "+wnd.Values()["id"]),
				ui.Text("Diese Seite hat ein Klick auf die Benachrichtigung geöffnet."),
			).Alignment(ui.Leading).Frame(ui.Frame{}.Larger())
		})
	}).Run()
}

func order(id string) push.Message {
	return push.Message{
		Title: "Neuer Auftrag " + id,
		Body:  "Ein Kunde wartet auf deine Antwort.",
		// a newer notification of the same order replaces the older one
		Tag:    "order-" + id,
		Route:  "order",
		Values: core.Values{"id": id},
	}
}

// sendLater sends to the browsers of the user in the background, so that the tab can be closed meanwhile.
func sendLater(wnd core.Window, uc push.UseCases) {
	subject := wnd.Subject()
	alert.ShowBannerMessage(wnd, alert.Message{Title: "Geplant", Message: "Die Benachrichtigung kommt in 10 Sekunden.", Intent: alert.IntentOk})

	go func() {
		time.Sleep(10 * time.Second)
		report, err := uc.SendToMe(subject, order("7"))
		if err != nil {
			slog.Error("tutorial: cannot send", "err", err)
			return
		}

		slog.Info("tutorial: sent", "report", describe(report))
	}()
}

func describe(r push.Report) string {
	return fmt.Sprintf("zugestellt an %d Gerät(e), entfernt: %d, fehlgeschlagen: %d", r.Sent, r.Removed, r.Failed)
}
