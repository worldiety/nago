// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"io"
	"net/http"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
)

// A logout and a login as another user in the same browser keep the cookie. A handler registered with
// HandleFuncSubject must see the change immediately and never serve the previous user.
func TestHandleFuncSubjectFollowsLoginAndLogout(t *testing.T) {
	var cfg *application.Configurator
	base := nagotest.Serve(t, func(c *application.Configurator) {
		cfg = c
		c.SetApplicationID("de.worldiety.subjecthandlertest")
		if err := c.HandleFuncSubject("/api/whoami", func(w http.ResponseWriter, r *http.Request, subject auth.Subject) {
			if !subject.Valid() {
				_, _ = io.WriteString(w, "anonymous")
				return
			}

			_, _ = io.WriteString(w, string(subject.ID()))
		}); err != nil {
			t.Fatal(err)
		}
	})

	users := std.Must(cfg.UserManagement()).UseCases
	sessions := std.Must(cfg.SessionManagement()).UseCases

	create := func(mail string) user.ID {
		t.Helper()
		usr, err := users.Create(user.SU(), user.ShortRegistrationUser{Firstname: "A", Lastname: "B", Email: user.Email(mail), Password: "Sup3r-Geheim!2026", PasswordRepeated: "Sup3r-Geheim!2026", Verified: true})
		if err != nil {
			t.Fatal(err)
		}
		return usr.ID
	}

	alice := create("alice@example.com")
	bob := create("bob@example.com")

	const sid = session.ID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	whoami := func() string {
		t.Helper()
		req, _ := http.NewRequest(http.MethodGet, base+"/api/whoami", nil)
		req.AddCookie(&http.Cookie{Name: "wdy-ora-access", Value: string(sid)})
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		buf, _ := io.ReadAll(res.Body)
		return string(buf)
	}

	if got := whoami(); got != "anonymous" {
		t.Fatalf("expected anonymous, got %q", got)
	}

	if err := sessions.LoginUser(sid, alice); err != nil {
		t.Fatal(err)
	}

	if got := whoami(); got != string(alice) {
		t.Fatalf("expected alice, got %q", got)
	}

	if _, err := sessions.Logout(sid); err != nil {
		t.Fatal(err)
	}

	if got := whoami(); got != "anonymous" {
		t.Fatalf("a logged out session must be anonymous, got %q", got)
	}

	if err := sessions.LoginUser(sid, bob); err != nil {
		t.Fatal(err)
	}

	if got := whoami(); got != string(bob) {
		t.Fatalf("expected bob and never alice, got %q", got)
	}
}
