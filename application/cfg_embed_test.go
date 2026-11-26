// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// A path of EmbeddablePages matches like the id of a root view: a wildcard all pages below it, any other path only
// the same page.
func TestEmbeddablePagePaths(t *testing.T) {
	pages := []proto.RootViewID{"e/*", "x/page"}
	for _, row := range []struct {
		path string
		want bool
	}{
		{path: "/e/forms/contact", want: true},
		{path: "/e/a", want: true},
		{path: "/e", want: false},
		{path: "/export/a", want: false},
		{path: "/v/forms/contact", want: false},
		{path: "/x/page", want: true},
		{path: "/x/page/below", want: false},
		{path: "/x/pages", want: false},
		{path: "/", want: false},
	} {
		if got := embeddable(pages, row.path); got != row.want {
			t.Errorf("%s: embeddable %v, want %v", row.path, got, row.want)
		}
	}

	if embeddable(nil, "/e/forms/contact") {
		t.Error("without embeddable pages, no page is embeddable")
	}
}

// EmbeddablePages refuses a path, which is no valid id of a root view, like RootView.
func TestEmbeddablePagesRefusesInvalidPaths(t *testing.T) {
	for _, row := range []struct {
		path  core.NavigationPath
		valid bool
	}{
		{path: "e/*", valid: true},
		{path: "x/page", valid: true},
		{path: ".", valid: true},
		{path: "e/", valid: false},
		{path: "/e/*", valid: false},
	} {
		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("%v", r)
				}
			}()

			(&Configurator{}).EmbeddablePages(row.path)
			return nil
		}()

		if (err == nil) != row.valid {
			t.Errorf("%q: refused %v, want %v", row.path, err, !row.valid)
		}
	}
}

// The session of a request comes from the plain cookie, see sessionIDFromCookies. Without one, the embedded cookie
// counts, but only under the name, which a response to the same request issues, and only with an id, which nago
// created for an embedded page.
func TestSessionIDOfEmbeddedCookie(t *testing.T) {
	embedded := string(session.NewEmbeddedID())
	for _, row := range []struct {
		name    string
		policy  cookiePolicy
		https   bool
		host    string // empty is localhost
		cookies string
		want    string
	}{
		{name: "https, embedded", https: true, cookies: "__Host-wdy-ora-embedded=" + embedded, want: embedded},
		{name: "https, both", https: true, cookies: "__Host-wdy-ora-access=anon; __Host-wdy-ora-embedded=" + embedded, want: "anon"},
		{name: "https, both, the other order", https: true, cookies: "__Host-wdy-ora-embedded=" + embedded + "; __Host-wdy-ora-access=anon", want: "anon"},
		{name: "https, tossed plain cookie and embedded", https: true, cookies: "wdy-ora-access=anon; __Host-wdy-ora-embedded=" + embedded, want: embedded},
		{name: "https, embedded cookie of a plain id", https: true, cookies: "__Host-wdy-ora-embedded=user1", want: ""},
		{name: "https, the embedded cookie of localhost", https: true, cookies: "wdy-ora-embedded=" + embedded, want: ""},
		{name: "localhost, embedded", cookies: "wdy-ora-embedded=" + embedded, want: embedded},
		{name: "localhost, both", cookies: "wdy-ora-access=anon; wdy-ora-embedded=" + embedded, want: "anon"},
		{name: "localhost, the embedded cookie of https", cookies: "__Host-wdy-ora-embedded=" + embedded, want: ""},
		{name: "127.0.0.1, embedded", host: "127.0.0.1:3000", cookies: "wdy-ora-embedded=" + embedded, want: embedded},
		{name: "plain http of another host, embedded", host: "app.example.com", cookies: "wdy-ora-embedded=" + embedded, want: ""},
		{name: "plain http of another host, plain", host: "app.example.com", cookies: "wdy-ora-access=anon", want: "anon"},
		{name: "NO_SSL=true, localhost, embedded", policy: cookieNeverSecure, cookies: "wdy-ora-embedded=" + embedded, want: ""},
		{name: "NO_SSL=true, https, embedded", policy: cookieNeverSecure, https: true, cookies: "__Host-wdy-ora-embedded=" + embedded, want: ""},
		{name: "NO_SSL=true, localhost, plain", policy: cookieNeverSecure, cookies: "wdy-ora-access=anon", want: "anon"},
		{name: "NO_SSL=false, plain http of another host, embedded", policy: cookieAlwaysSecure, host: "app.example.com", cookies: "__Host-wdy-ora-embedded=" + embedded, want: embedded},
	} {
		r := httptest.NewRequest(http.MethodGet, "/wire", nil)
		r.Host = "localhost:3000"
		if row.host != "" {
			r.Host = row.host
		}

		if row.https {
			r.Header.Set("X-Forwarded-Proto", "https")
		}

		r.Header.Set("Cookie", row.cookies)
		c := &Configurator{cookiePolicy: row.policy}
		id, migrate := c.sessionIDOf(r)
		if id != row.want || migrate {
			t.Errorf("%s: got %q, migrate %v, want %q", row.name, id, migrate, row.want)
		}
	}
}
