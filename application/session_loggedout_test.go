// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// A logout, e.g. in another tab or by a failed single sign-on refresh in the background, shows all open windows of
// the session as logged out at once, but no window of another session.
func TestLogoutReachesTheOpenWindowsOfTheSession(t *testing.T) {
	var cfg *application.Configurator
	app := nagotest.New(t, func(c *application.Configurator) {
		cfg = c
		c.SetApplicationID("de.worldiety.loggedouttest")
		std.Must(c.SessionManagement())
		c.RootView(".", func(wnd core.Window) core.View {
			if wnd.Subject().Valid() {
				return ui.Text("signed in")
			}

			return ui.Text("anonymous")
		})
	})

	users := std.Must(cfg.UserManagement()).UseCases
	sessions := std.Must(cfg.SessionManagement()).UseCases
	login := func(sid session.ID, mail string) {
		t.Helper()
		usr, err := users.Create(user.SU(), user.ShortRegistrationUser{Firstname: "A", Lastname: "B", Email: user.Email(mail), Password: "Sup3r-Geheim!2026", PasswordRepeated: "Sup3r-Geheim!2026", Verified: true})
		if err != nil {
			t.Fatal(err)
		}

		if err := sessions.LoginUser(sid, usr.ID); err != nil {
			t.Fatal(err)
		}
	}

	alice := session.ID(strings.Repeat("a", 40))
	bob := session.ID(strings.Repeat("b", 40))
	login(alice, "alice@example.com")
	login(bob, "bob@example.com")

	tab1 := app.Open(t, nil, ".", nagotest.Session(string(alice)))
	tab2 := app.Open(t, nil, ".", nagotest.Session(string(alice)))
	other := app.Open(t, nil, ".", nagotest.Session(string(bob)))
	for _, w := range []*nagotest.Window{tab1, tab2, other} {
		w.Find(nagotest.Text("signed in")).Exactly(1)
	}

	if _, err := sessions.Logout(alice); err != nil {
		t.Fatal(err)
	}

	tab1.WaitFor(nagotest.Text("anonymous"), 5*time.Second)
	tab2.WaitFor(nagotest.Text("anonymous"), 5*time.Second)

	other.Settle()
	other.Find(nagotest.Text("signed in")).Exactly(1)
}
