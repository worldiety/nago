// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgonboarding_test

import (
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/onboarding"
	cfgonboarding "go.wdy.de/nago/application/onboarding/cfg"
	uionboarding "go.wdy.de/nago/application/onboarding/ui"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type mailbox struct {
	mutex sync.Mutex
	codes []string
}

func (m *mailbox) deliver(to user.Email, code string, validUntil time.Time) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.codes = append(m.codes, code)
	return nil
}

func (m *mailbox) last(t *testing.T) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if len(m.codes) == 0 {
		t.Fatal("no code has been sent")
	}

	return m.codes[len(m.codes)-1]
}

func setupApp(t *testing.T, email string, box *mailbox) *nagotest.App {
	return nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.onboardingtest")
		cfg.SetName("Testapp")
		cfg.RootView(".", func(wnd core.Window) core.View {
			if wnd.Subject().Valid() {
				return ui.Text("home of " + wnd.Subject().Name())
			}

			return ui.Text("home")
		})

		cfg.RootView("impressum", func(wnd core.Window) core.View {
			return ui.Text("legal notice")
		})

		if _, err := cfgonboarding.Enable(cfg, cfgonboarding.Options{Email: email, Deliver: box.deliver, Exempt: []core.NavigationPath{"impressum"}}); err != nil {
			t.Fatal(err)
		}

		// a second invocation returns the installed module
		if _, err := cfgonboarding.Enable(cfg, cfgonboarding.Options{}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSetupFirstUser(t *testing.T) {
	box := &mailbox{}
	app := setupApp(t, "torben@example.com", box)
	sid := strings.Repeat("a", 32)

	// every route shows the setup, except the exempted ones
	app.Open(t, nil, "impressum").Find(nagotest.Text("legal notice")).Exactly(1)
	w := app.Open(t, nil, ".", nagotest.Session(sid))
	w.Find(nagotest.TextContains("t***@example.com")).Exactly(1)
	w.FindAll(nagotest.Text("home")).None()

	w.Click(w.Find(nagotest.Text("Einrichtung starten")))
	w.Type(w.Find(nagotest.ID("onboarding-code")), "000000x")
	w.Click(w.Find(nagotest.Text("Bestätigen")))
	w.Find(nagotest.ID("onboarding-code")).Exactly(1) // still the code step

	w.Type(w.Find(nagotest.ID("onboarding-code")), box.last(t))
	w.Click(w.Find(nagotest.Text("Bestätigen")))

	// the setup offers the single sign-on first, the user goes back and forth
	w.Find(nagotest.TextContains("login.worldiety.nago.app")).Exactly(1)
	w.Click(w.Find(nagotest.Text("Mit E-Mail und Passwort")))
	w.Click(w.Find(nagotest.Text("Zurück")))
	w.Click(w.Find(nagotest.Text("Mit E-Mail und Passwort")))

	w.Type(w.Find(nagotest.ID("onboarding-firstname")), "Torben")
	w.Type(w.Find(nagotest.ID("onboarding-lastname")), "Schinke")
	w.Type(w.Find(nagotest.ID("onboarding-password")), "Sup3r-Geheim!2026")
	w.Type(w.Find(nagotest.ID("onboarding-password-repeated")), "Sup3r-Geheim!2026")
	w.Click(w.Find(nagotest.Text("Konto anlegen")))
	// the session is logged in and the instance is set up
	w.WaitFor(nagotest.TextContains("home of"), 5*time.Second)

	users := std.Must(app.Configurator().UserManagement())
	if n, err := users.UseCases.CountUsers(); err != nil || n != 1 {
		t.Fatalf("expected exactly one user, got %d %v", n, err)
	}

	usr := std.Must(users.UseCases.FindByMail(user.SU(), "torben@example.com")).Unwrap()
	if !usr.EMailVerified || usr.RequiresVerification() || usr.Contact.Firstname != "Torben" {
		t.Fatalf("the user has not been set up: %+v", usr)
	}

	// other visitors see the normal application
	app.Open(t, nil, ".").Find(nagotest.Text("home")).Exactly(1)
	app.Open(t, nil, "onboarding/setup").Find(nagotest.Text("Diese Anwendung ist bereits eingerichtet.")).Exactly(1)
}

// The first user chooses the single sign-on and who else may sign in. Until the first user has signed in by the
// login service, only its address may.
func TestSetupFirstUserWithSSO(t *testing.T) {
	box := &mailbox{}
	app := setupApp(t, "torben@example.com", box)
	w := app.Open(t, nil, ".", nagotest.Session(strings.Repeat("b", 32)))
	w.Click(w.Find(nagotest.Text("Einrichtung starten")))
	w.Type(w.Find(nagotest.ID("onboarding-code")), box.last(t))
	w.Click(w.Find(nagotest.Text("Bestätigen")))
	w.Click(w.Find(nagotest.Text("Mit dem Nago-Login (SSO)")))

	// the domain of the address is offered and preselected
	w.Find(nagotest.Text("Alle von example.com")).Exactly(1)
	w.Click(w.Find(nagotest.Text("Ich und ausgewählte Personen oder Firmen")))
	w.Type(w.Find(nagotest.ID("onboarding-sso-entries")), "Anna Muster <Anna@Firma.de>; firma2.de, gmail.com")
	w.Find(nagotest.TextContains("öffentlicher Mail-Anbieter")).Exactly(1)
	w.Find(nagotest.Text("anna@firma.de")).Exactly(1)
	w.Find(nagotest.Text("alle Adressen von firma2.de")).Exactly(1)
	w.Click(w.Find(nagotest.Text("Weiter zum Nago-Login")))
	if len(w.Links()) != 0 {
		t.Fatalf("a public provider must stop the setup: %v", w.Links())
	}

	w.Type(w.Find(nagotest.ID("onboarding-sso-entries")), "Anna Muster <Anna@Firma.de>; firma2.de")
	w.FindAll(nagotest.TextContains("öffentlicher Mail-Anbieter")).None()
	w.Click(w.Find(nagotest.Text("Weiter zum Nago-Login")))

	// the window continues at the login service
	links := w.Links()
	if len(links) != 1 || !strings.HasPrefix(links[0], onboarding.DefaultNLSServer+"?") || !strings.Contains(links[0], "nonce=") {
		t.Fatalf("expected the login service, got %v", links)
	}

	users := std.Must(app.Configurator().UserManagement())
	usr := std.Must(users.UseCases.FindByMail(user.SU(), "torben@example.com")).Unwrap()
	if !usr.EMailVerified || !usr.RequirePasswordChange {
		t.Fatalf("the user has not been set up without a password: %+v", usr)
	}

	loadGlobal := std.Must(app.Configurator().SettingsManagement()).UseCases.LoadGlobal
	usrSettings := settings.ReadGlobal[user.Settings](loadGlobal)
	if usrSettings.SSONLSServer != onboarding.DefaultNLSServer || !slices.Equal(usrSettings.SSOAllowList, []string{`^torben@example\.com$`}) {
		t.Fatalf("only the first user may sign in by now: %+v", usrSettings)
	}

	// the first user has signed in by the login service, which merges the account
	if _, err := users.UseCases.MergeSingleSignOnUser(user.SingleSignOnUser{Email: "torben@example.com", Firstname: "Torben"}, nil); err != nil {
		t.Fatal(err)
	}

	app.Configurator().EventBus().Publish(session.Authenticated{Session: "x", User: usr.ID})
	want := []string{`^torben@example\.com$`, `^anna@firma\.de$`, `^[^@]+@firma2\.de$`}
	deadline := time.Now().Add(5 * time.Second)
	for !slices.Equal(settings.ReadGlobal[user.Settings](loadGlobal).SSOAllowList, want) {
		if time.Now().After(deadline) {
			t.Fatalf("expected %v, got %v", want, settings.ReadGlobal[user.Settings](loadGlobal).SSOAllowList)
		}

		time.Sleep(10 * time.Millisecond)
	}

	app.Open(t, nil, ".").Find(nagotest.Text("home")).Exactly(1)
}

func TestSetupWithoutEmail(t *testing.T) {
	t.Setenv(onboarding.EnvInitialUserEmail, "")
	t.Setenv(onboarding.EnvHubOnboardingMail, "")
	app := setupApp(t, "", &mailbox{})
	w := app.Open(t, nil, ".")
	w.Find(nagotest.ID(uionboarding.IDProblem)).Find(nagotest.Text("Für diese Anwendung ist keine gültige Adresse für das erste Konto eingetragen.")).Exactly(1)
	w.FindAll(nagotest.Text("Einrichtung starten")).None()
}
