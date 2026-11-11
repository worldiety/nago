// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Command tutorial-121-self-registration shows the self-service account creation: anonymous visitors find a
// register link on the login page and enter their mail address. The mailed link opens the rest of the
// registration with the contact, the password and a last step with the consents.
//
// Without an SMTP server or the Nago Mail Service, the mail stays in the outgoing queue: log in as the bootstrap
// admin and open admin/mail/outgoing to find the link. Start it with NAGO_MAIL_SERVICE=off to not use the
// service at all.
package main

import (
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial.selfregistration")
		cfg.Serve(vuejs.Dist())
		cfg.SetName("Tutorial")

		std.Must(cfg.Authentication())
		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		option.MustZero(cfg.StandardSystems())

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		configureSelfRegistration(cfg)

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			if !wnd.Subject().Valid() {
				return ui.VStack(
					ui.Text("Willkommen! Bitte anmelden oder ein Konto erstellen."),
					ui.PrimaryButton(func() {
						wnd.Navigation().ForwardTo("account/register", nil)
					}).Title("Konto erstellen"),
				).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
			}

			return ui.VStack(
				ui.Text("Hallo " + wnd.Subject().Name()),
			).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}

func configureSelfRegistration(cfg *application.Configurator) {
	settingsMgmt := std.Must(cfg.SettingsManagement())
	usrSettings := settings.ReadGlobal[user.Settings](settingsMgmt.UseCases.LoadGlobal)

	usrSettings.SelfRegistration = true
	usrSettings.SelfPasswordReset = true

	// "" hides a field, ^.*$ makes it optional, ^.+$ required, anything else validates by regex
	usrSettings.Salutation = "^(Frau|Herr|Divers)$"
	usrSettings.CompanyName = "^.+$"
	usrSettings.Position = "^.*$"
	usrSettings.MobilePhone = "^.*$"

	usrSettings.Consents = []user.ConsentOption{
		{
			ID:       consent.DataProtectionProvision,
			Register: user.ConsentText{Label: "Ja, ich habe die [Datenschutzbestimmungen](/page/datenschutz) gelesen und akzeptiert."},
			Required: true,
		},
		{
			ID:       consent.GeneralTermsAndConditions,
			Register: user.ConsentText{Label: "Ja, ich habe die [Geschäftsbedingungen](/page/agb) gelesen und akzeptiert."},
			Required: true,
		},
		{
			ID: consent.Newsletter,
			Register: user.ConsentText{
				Label:          "Ja, ich melde mich zum Newsletter an. Eine Abbestellung ist jederzeit möglich.",
				SupportingText: "Ein Widerspruch ist jederzeit in den Einstellungen Ihres Benutzerkontos möglich.",
			},
		},
	}

	option.MustZero(settingsMgmt.UseCases.StoreGlobal(user.SU(), usrSettings))
}
