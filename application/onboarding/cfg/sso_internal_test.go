// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgonboarding

import (
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// The access, which the first user has chosen for the single sign-on, applies only after it has signed in by the
// login service. A sign-in by password keeps the gate, which only allows the first user.
func TestApplyOnSignIn(t *testing.T) {
	gate := []string{`^torben@example\.com$`}
	domain := []string{`^[^@]+@example\.com$`}
	for _, tc := range []struct {
		name    string
		nls     bool // the first user has signed in by the login service
		someone bool // another user has signed in
		want    []string
	}{
		{name: "password", want: gate},
		{name: "login service", nls: true, want: domain},
		{name: "unknown user", nls: true, someone: true, want: gate},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var mutex sync.Mutex
			var codes []string
			var management Management
			app := nagotest.New(t, func(cfg *application.Configurator) {
				cfg.SetApplicationID("de.worldiety.onboardingtest")
				cfg.SetName("Testapp")
				cfg.RootView(".", func(wnd core.Window) core.View { return ui.Text("home") })
				management = std.Must(Enable(cfg, Options{Email: "torben@example.com", Deliver: func(to user.Email, code string, validUntil time.Time) error {
					mutex.Lock()
					defer mutex.Unlock()
					codes = append(codes, code)
					return nil
				}}))
			})

			// the first user chooses the single sign-on for everyone of its domain
			w := app.Open(t, nil, ".", nagotest.Session(strings.Repeat("e", 32)))
			w.Click(w.Find(nagotest.Text("Einrichtung starten")))
			mutex.Lock()
			code := codes[len(codes)-1]
			mutex.Unlock()
			w.Type(w.Find(nagotest.ID("onboarding-code")), code)
			w.Click(w.Find(nagotest.Text("Bestätigen")))
			w.Click(w.Find(nagotest.Text("Mit dem Nago-Login (SSO)")))
			w.Click(w.Find(nagotest.Text("Weiter zum Nago-Login")))

			users := std.Must(app.Configurator().UserManagement()).UseCases
			loadGlobal := std.Must(app.Configurator().SettingsManagement()).UseCases.LoadGlobal
			if got := settings.ReadGlobal[user.Settings](loadGlobal).SSOAllowList; !slices.Equal(got, gate) {
				t.Fatalf("only the first user may sign in by now: %v", got)
			}

			signedIn := std.Must(users.FindByMail(user.SU(), "torben@example.com")).Unwrap().ID
			if tc.nls {
				// the login service merges the account at the first sign-in
				if _, err := users.MergeSingleSignOnUser(user.SingleSignOnUser{Email: "torben@example.com", Firstname: "Torben"}, nil); err != nil {
					t.Fatal(err)
				}
			}

			if tc.someone {
				signedIn = "someone"
			}

			applyOnSignIn(users.FindByID, management.UseCases.ApplySSOAccess)(session.Authenticated{Session: "s", User: signedIn})
			if got := settings.ReadGlobal[user.Settings](loadGlobal).SSOAllowList; !slices.Equal(got, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
