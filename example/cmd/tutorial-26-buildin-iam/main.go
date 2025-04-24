// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// #[go.permission.generateTable]
package main

import (
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	cfgusercircle "go.wdy.de/nago/application/usercircle/cfg"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xreflect"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

var myPermission = permission.Declare[SayHello]("de.worldiety.tutorial.say_hello", "Jeden Grüßen", "Diese Erlaubnis muss dem Nutzer zugewiesen werden.")

// SayHello greets everyone who has been authenticated.
type SayHello func(auth auth.Subject) string

func NewSayHello() SayHello {
	return func(auth auth.Subject) string {
		if err := auth.Audit(myPermission); err != nil {
			return fmt.Sprintf("invalid: %v", err)
		}

		return "hello " + auth.Name()
	}
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())
		cfg.SetName("Tutorial")

		std.Must(cfg.Authentication())
		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		option.MustZero(cfg.StandardSystems())
		option.Must(cfgusercircle.Enable(cfg))

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		configureGDPRConsents(cfg)

		sayHello := NewSayHello()

		// remember to update your user.Settings to use the matching regex like ^(Kaufmann|Informatiker)$
		xreflect.SetFieldTagFor[user.Settings]("ProfessionalGroup", "supportingText", "Die Berufsgruppe muss Kaufmann oder Informatiker sein.")
		xreflect.SetFieldTagFor[user.Settings]("MobilePhone", "supportingText", "Wir benötigen die Telefonnummer, um Ihnen Buchungsbestätigungen zu schicken.")

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text(fmt.Sprintf("%s", sayHello(wnd.Subject()))),
			).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
		})

	}).Run()
}

func configureGDPRConsents(cfg *application.Configurator) {
	usrSettings := settings.ReadGlobal[user.Settings](std.Must(cfg.SettingsManagement()).UseCases.LoadGlobal)

	// do not append, just clear it
	usrSettings.Consents = []user.ConsentOption{
		{
			ID:       consent.DataProtectionProvision,
			Register: user.ConsentText{Label: "Ja, ich habe die [Datenschutzbestimmungen](/page/datenschutz) gelesen und akzeptiert."},
			Required: true,
		},
		{
			ID: consent.Newsletter,
			Register: user.ConsentText{
				Label:          "Ja, ich melde mich zum Newsletter an. Eine Abbestellung ist jederzeit möglich.",
				SupportingText: "Ein Widerspruch ist jederzeit in den Einstellungen Ihres Benutzerkontos/über Abmeldelink in den E-Mails möglich, ohne dass weitere (Übermittlungs-)Kosten als die nach den Basistarifen entstehen.",
			},
			Profile: user.ConsentText{
				Label:          "Newsletter",
				SupportingText: "Regelmäßig Email Updates erhalten",
			},
		},
		{
			ID:       consent.GeneralTermsAndConditions,
			Register: user.ConsentText{Label: "Ja, ich habe die [Geschäftsbedingungen](/page/agb) gelesen und akzeptiert."},
			Required: true,
		},
		{
			ID: consent.TermsOfUse,
			Register: user.ConsentText{
				Label: "Ja, ich habe die [Nutzungsbedingungen](/page/nutzungsbedingungen) gelesen und akzeptiert.",
			},
			Required: true,
		},
		{
			ID: consent.MinAge,
			Register: user.ConsentText{
				Label: "Ja, ich bestätige, dass ich mindestens 16 Jahre alt bin.",
			},
			Required: true,
		},
		{
			ID: consent.SMS,
			Register: user.ConsentText{
				Label:          "Ja, ich melde mich zum SMS Versand an, um kurzfristige Benachrichtigungen zu erhalten. Eine Abbestellung ist jederzeit möglich.",
				SupportingText: "Ein Widerspruch ist jederzeit in den Einstellungen Ihres Benutzerkontos möglich.",
			},
			Profile: user.ConsentText{
				Label:          "SMS Versand",
				SupportingText: "kurzfristige Updates zu Veranstaltungen erhalten.",
			},
			Required: false,
		},
		{
			ID: "my.custom.consent",
			Register: user.ConsentText{
				Label: "Accept something completely different.",
			},
			Profile: user.ConsentText{
				Label: "Accept something completely different.",
			},
			Required: false,
		},
	}

	// apply settings
	option.MustZero(option.Must(cfg.SettingsManagement()).UseCases.StoreGlobal(user.SU(), usrSettings))
}
