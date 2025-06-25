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
	"go.wdy.de/nago/application/consent"
	cfgdataimport "go.wdy.de/nago/application/dataimport/cfg"
	"go.wdy.de/nago/application/dataimport/importer/userimporter"
	"go.wdy.de/nago/application/dataimport/parser/csv"
	"go.wdy.de/nago/application/dataimport/parser/json"
	"go.wdy.de/nago/application/dataimport/parser/pdf"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_62")

		cfg.Serve(vuejs.Dist())
		cfg.SetDecorator(cfg.NewScaffold().
			Decorator())

		option.MustZero(cfg.StandardSystems())
		users := std.Must(cfg.UserManagement()).UseCases
		std.Must(users.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		option.Must(cfginspector.Enable(cfg))

		imports := std.Must(cfgdataimport.Enable(cfg))
		option.MustZero(imports.UseCases.RegisterImporter(user.SU(), userimporter.NewImporter(users)))
		option.MustZero(imports.UseCases.RegisterParser(user.SU(), csv.NewParser()))
		option.MustZero(imports.UseCases.RegisterParser(user.SU(), pdf.NewParser()))
		option.MustZero(imports.UseCases.RegisterParser(user.SU(), json.NewParser()))

		configureGDPRConsents(cfg)

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text("importer example"),
			).Frame(ui.Frame{}.MatchScreen())

		})
	}).
		Run()
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
				Label: "Ja, ich habe die [Nutzungsbedingungen](http://localhost:3000/admin/data/entry?stage=95ad3e6442a7c7de61f8d5b70ae38cd4&importer=nago.data.importer.user&entry=95ad3e6442a7c7de61f8d5b70ae38cd4%2F1750422709492%2F++++2) gelesen und akzeptiert.",
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
