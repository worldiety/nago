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
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

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

// Over https, only the __Host- cookie counts, because another application of the same site may have set the plain
// one. The plain cookie of a logged in session of an older version is taken over once, by the wire or the page.
func TestHandleFuncSubjectOverHTTPS(t *testing.T) {
	var cfg *application.Configurator
	base := nagotest.Serve(t, func(c *application.Configurator) {
		cfg = c
		c.SetApplicationID("de.worldiety.subjecthandlerhttpstest")
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

	usr, err := users.Create(user.SU(), user.ShortRegistrationUser{Firstname: "A", Lastname: "B", Email: "alice@example.com", Password: "Sup3r-Geheim!2026", PasswordRepeated: "Sup3r-Geheim!2026", Verified: true})
	if err != nil {
		t.Fatal(err)
	}

	const alice = session.ID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	const tossed = session.ID("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
	if err := sessions.LoginUser(alice, usr.ID); err != nil {
		t.Fatal(err)
	}

	whoami := func(cookie string) string {
		t.Helper()
		req, _ := http.NewRequest(http.MethodGet, base+"/api/whoami", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("Cookie", cookie)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		buf, _ := io.ReadAll(res.Body)
		return string(buf)
	}

	for name, tc := range map[string]struct {
		cookie string
		want   string
	}{
		"https cookie":                  {cookie: "__Host-wdy-ora-access=" + string(alice), want: string(usr.ID)},
		"tossed next to the own cookie": {cookie: "wdy-ora-access=" + string(tossed) + "; __Host-wdy-ora-access=" + string(alice), want: string(usr.ID)},
		"tossed anonymous session":      {cookie: "wdy-ora-access=" + string(tossed), want: "anonymous"},
		"logged in by an older version": {cookie: "wdy-ora-access=" + string(alice), want: string(usr.ID)},
		"ambiguous plain cookies":       {cookie: "wdy-ora-access=" + string(tossed) + "; wdy-ora-access=" + string(alice), want: "anonymous"},
	} {
		if got := whoami(tc.cookie); got != tc.want {
			t.Errorf("%s: got %q, want %q", name, got, tc.want)
		}
	}

	// the wire takes the session of an older version over into the __Host- cookie and removes the plain one
	u, _ := url.Parse(base)
	u.Scheme = "ws"
	u.Path = "/wire"
	u.RawQuery = url.Values{"_sid": {strings.Repeat("m", 40)}}.Encode()
	header := http.Header{}
	header.Set("X-Forwarded-Proto", "https")
	header.Set("Cookie", "wdy-ora-access="+string(alice))
	conn, res, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	got := map[string]*http.Cookie{}
	for _, c := range res.Cookies() {
		got[c.Name] = c
	}

	if c := got["__Host-wdy-ora-access"]; c == nil || c.Value != string(alice) || !c.Secure || c.Path != "/" {
		t.Fatalf("expected the session to be taken over into the __Host- cookie, got %+v", c)
	}

	if c := got["wdy-ora-access"]; c == nil || c.MaxAge >= 0 {
		t.Fatalf("expected the plain cookie to be removed, got %+v", c)
	}
}
