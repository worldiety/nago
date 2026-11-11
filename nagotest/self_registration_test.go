// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest_test

import (
	"html"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/mail"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

var registerLink = regexp.MustCompile(`account/register\?[^"\s<]+`)

// registrationLink returns the values of the last registration link in the mail queue.
func registrationLink(t *testing.T, cfg *application.Configurator) core.Values {
	t.Helper()

	mails, err := cfg.MailManagement()
	if err != nil {
		t.Fatal(err)
	}

	var link string
	for id, err := range mails.UseCases.FindOutgoingIDs(user.SU(), mail.OutgoingFilter{}) {
		if err != nil {
			t.Fatal(err)
		}

		optMail, err := mails.UseCases.FindOutgoingByID(user.SU(), id)
		if err != nil || optMail.IsNone() {
			t.Fatal("cannot find mail", err)
		}

		for _, part := range optMail.Unwrap().Mail.Parts {
			if m := registerLink.FindString(string(part.Encoded)); m != "" {
				// a fresh instance has the template, older ones get the plain text fallback
				if !strings.Contains(part.Header.Get("Content-Type"), "text/html") {
					t.Fatalf("expected the html template, got %v", part.Header)
				}

				link = html.UnescapeString(m)
			}
		}
	}

	if link == "" {
		t.Fatal("no registration link has been sent")
	}

	u, err := url.Parse("http://localhost/" + link)
	if err != nil {
		t.Fatal(err)
	}

	return core.Values{"id": u.Query().Get("id"), "code": u.Query().Get("code")}
}

// assertSteps checks the compact list of steps, which tells the visitor how many steps are left.
func assertSteps(t *testing.T, w *nagotest.Window, current int, headlines ...string) {
	t.Helper()

	s := w.Find(nagotest.Type[*proto.Stepper]()).Node().Component.(*proto.Stepper)
	if s.Layout != proto.StepperLayoutSimpleList || int(s.Value) != current || len(s.Steps) != len(headlines) {
		t.Fatalf("unexpected stepper: layout=%v value=%v steps=%d", s.Layout, s.Value, len(s.Steps))
	}

	for i, h := range headlines {
		if string(s.Steps[i].Title) != h {
			t.Fatalf("step %d: expected %q, got %q", i, h, s.Steps[i].Title)
		}
	}
}

func TestSelfRegistration(t *testing.T) {
	t.Setenv(application.EnvMailService, "off")

	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		if _, err := cfg.Authentication(); err != nil {
			t.Fatal(err)
		}

		sets, err := cfg.SettingsManagement()
		if err != nil {
			t.Fatal(err)
		}

		usrSettings := settings.ReadGlobal[user.Settings](sets.UseCases.LoadGlobal)
		usrSettings.SelfRegistration = true
		usrSettings.Consents = []user.ConsentOption{{
			ID:       consent.DataProtectionProvision,
			Register: user.ConsentText{Label: "Datenschutz akzeptiert"},
			Required: true,
		}}

		if err := sets.UseCases.StoreGlobal(user.SU(), usrSettings); err != nil {
			t.Fatal(err)
		}

		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.Text("Hallo " + wnd.Subject().Name())
		})
	})

	cfg := app.Configurator()

	// first step: only the mail address, the visitor sees all steps of the registration
	w := app.Open(t, nil, "account/register")
	assertSteps(t, w, 0, "E-Mail", "Kontakt", "Passwort", "Prüfen")
	w.Type(w.Find(nagotest.Label("E-Mail Adresse")), "erika@example.com")
	w.Click(w.Find(nagotest.Text("Link senden")))
	w.Find(nagotest.Text("Postfach prüfen"))

	// the address is highlighted like inline code, so that a typo stands out
	if font := w.Find(nagotest.Text("erika@example.com")).Node().Component.(*proto.TextView).Font; font.Name != proto.Str(ui.MonoFontName) {
		t.Fatalf("expected the address in a monospaced font, got %q", font.Name)
	}
	w.Find(nagotest.Text("Andere Adresse"))

	values := registrationLink(t, cfg)

	// a wrong code is rejected
	bad := app.Open(t, nil, "account/register", nagotest.Values(core.Values{"id": values["id"], "code": "wrong"}))
	bad.Find(nagotest.Text("Link nicht gültig"))

	// the link opens the rest of the registration with the confirmed address
	w = app.Open(t, nil, "account/register", nagotest.Values(values))
	w.Find(nagotest.Text("Bitte den Kontakt eingeben"))
	w.FindAll(nagotest.TextContains("ist bestätigt")).None()
	assertSteps(t, w, 1, "E-Mail", "Kontakt", "Passwort", "Prüfen")

	// the contact step validates before it can be left
	w.Click(w.Find(nagotest.Text("Weiter")))
	w.Find(nagotest.Text("Bitte den Kontakt eingeben"))

	w.Type(w.Find(nagotest.Label("Vorname*")), "Erika")
	w.Type(w.Find(nagotest.Label("Nachname*")), "Mustermann")
	w.Click(w.Find(nagotest.Text("Weiter")))
	w.Find(nagotest.Text("Bitte Passwort vergeben"))

	pwd := "Tutorial#Reg2026x"
	w.Type(w.Find(nagotest.Label("Passwort")), pwd)
	w.Type(w.Find(nagotest.Label("Passwort wiederholen")), pwd)
	w.Click(w.Find(nagotest.Text("Weiter")))
	w.Find(nagotest.Text("Fast geschafft..."))

	// the last step shows the data and the consents below
	w.Find(nagotest.Label("E-Mail Adresse"))
	w.FindAll(nagotest.Text("Sind die Daten korrekt?")).None()
	box := w.Find(nagotest.Type[*proto.Checkbox]()).Node().Component.(*proto.Checkbox)

	// the text of the consent is a label of the checkbox, so that a click on it toggles the checkbox
	var labelText strings.Builder
	for _, n := range w.FindAll(nagotest.Where("label of the checkbox", func(n nagotest.Node) bool {
		tv, ok := n.Component.(*proto.TextView)
		return ok && box.Id != "" && string(tv.LabelFor) == string(box.Id)
	})).Nodes() {
		labelText.WriteString(string(n.Component.(*proto.TextView).Value))
	}

	if !strings.Contains(labelText.String(), "Datenschutz akzeptiert") {
		t.Fatalf("expected the consent text to be a label of the checkbox %q, got %q", box.Id, labelText.String())
	}

	// a required consent blocks the creation
	w.Click(w.Find(nagotest.Text("Konto erstellen")))
	w.Find(nagotest.Text("Die Zustimmung ist erforderlich.")).Visible()

	w.Type(w.Find(nagotest.Type[*proto.Checkbox]()), "true")
	w.Click(w.Find(nagotest.Text("Konto erstellen")))
	w.Find(nagotest.Text("Hallo Erika Mustermann"))

	// the account is verified and the link is used up
	users, err := cfg.UserManagement()
	if err != nil {
		t.Fatal(err)
	}

	optUser, err := users.UseCases.FindByMail(user.SU(), "erika@example.com")
	if err != nil || optUser.IsNone() {
		t.Fatal("account not created", err)
	}

	if usr := optUser.Unwrap(); !usr.EMailVerified || usr.RequiresVerification() || usr.ConsentStatusByID(consent.DataProtectionProvision) != consent.Approved {
		t.Fatalf("unexpected account: verified=%v requiresVerification=%v", usr.EMailVerified, usr.RequiresVerification())
	}

	again := app.Open(t, nil, "account/register", nagotest.Values(values))
	again.Find(nagotest.Text("Link nicht gültig"))
}
