// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgonboarding_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	cfgonboarding "go.wdy.de/nago/application/onboarding/cfg"
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

func TestSetupWithoutEmail(t *testing.T) {
	app := setupApp(t, "", &mailbox{})
	w := app.Open(t, nil, ".")
	w.Find(nagotest.TextContains("Bitte wende dich an den Support")).Exactly(1)
	w.FindAll(nagotest.Text("Einrichtung starten")).None()
}
