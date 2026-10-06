// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRequestIsHTTPS(t *testing.T) {
	cases := []struct {
		name    string
		tls     bool
		headers map[string]string
		want    bool
	}{
		{name: "plain http in the local network", want: false},
		{name: "own tls", tls: true, want: true},
		{name: "proxy tells https", headers: map[string]string{"X-Forwarded-Proto": "https"}, want: true},
		{name: "proxy chain with https", headers: map[string]string{"X-Forwarded-Proto": "http, https"}, want: true},
		{name: "proxy tells http", headers: map[string]string{"X-Forwarded-Proto": "http", "X-Forwarded-For": "10.0.0.1"}, want: false},
		{name: "rfc 7239 https", headers: map[string]string{"Forwarded": `for=10.0.0.1;proto=https`}, want: true},
		{name: "rfc 7239 http", headers: map[string]string{"Forwarded": `for=10.0.0.1;proto="http"`}, want: false},
		{name: "rfc 7239 without proto", headers: map[string]string{"Forwarded": `for=10.0.0.1`}, want: true},
		{name: "proxy without protocol", headers: map[string]string{"X-Forwarded-For": "10.0.0.1"}, want: true},
		{name: "nginx real ip", headers: map[string]string{"X-Real-IP": "10.0.0.1"}, want: true},
	}

	for _, tc := range cases {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		if tc.tls {
			r.TLS = &tls.ConnectionState{}
		}

		for k, v := range tc.headers {
			r.Header.Set(k, v)
		}

		if got := requestIsHTTPS(r); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestCookiePolicy(t *testing.T) {
	plain := httptest.NewRequest(http.MethodGet, "/", nil)
	proxied := httptest.NewRequest(http.MethodGet, "/", nil)
	proxied.Header.Set("X-Forwarded-Proto", "https")

	for _, tc := range []struct {
		noSSL      string // empty means unset
		plain      bool
		proxied    bool
		wantPolicy cookiePolicy
	}{
		{noSSL: "", plain: false, proxied: true, wantPolicy: cookieAdaptive},
		{noSSL: "true", plain: false, proxied: false, wantPolicy: cookieNeverSecure},
		{noSSL: "false", plain: true, proxied: true, wantPolicy: cookieAlwaysSecure},
	} {
		if tc.noSSL != "" {
			t.Setenv(envNoSSL, tc.noSSL)
		}

		c := &Configurator{cookiePolicy: determineCookiePolicy()}
		if c.cookiePolicy != tc.wantPolicy || c.secureCookie(plain) != tc.plain || c.secureCookie(proxied) != tc.proxied {
			t.Errorf("NO_SSL=%q: policy %v, plain %v, proxied %v", tc.noSSL, c.cookiePolicy, c.secureCookie(plain), c.secureCookie(proxied))
		}

		wantName := sessionCookieName
		if tc.proxied {
			wantName = hostSessionCookieName
		}

		cookie := c.newSessionCookie(proxied, "id")
		if cookie.Name != wantName || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" || cookie.Secure != tc.proxied {
			t.Errorf("NO_SSL=%q: unexpected cookie %+v", tc.noSSL, cookie)
		}
	}
}

// Another application of the same site can set a plain cookie for the parent domain, but no __Host- cookie. Over
// https, the plain cookie is taken over only once from an older version, if it is unambiguous and logged in.
func TestSessionIDFromCookies(t *testing.T) {
	loggedIn := func(id string) bool { return strings.HasPrefix(id, "user") }
	for _, tc := range []struct {
		name        string
		secure      bool
		cookies     string
		wantID      string
		wantMigrate bool
	}{
		{name: "plain http", cookies: "wdy-ora-access=anon", wantID: "anon"},
		{name: "plain http ignores the https cookie", cookies: "__Host-wdy-ora-access=user1", wantID: ""},
		{name: "https", secure: true, cookies: "__Host-wdy-ora-access=anon", wantID: "anon"},
		{name: "https ignores a tossed cookie", secure: true, cookies: "wdy-ora-access=user-evil; __Host-wdy-ora-access=user1", wantID: "user1"},
		{name: "https takes over a logged in session once", secure: true, cookies: "wdy-ora-access=user1", wantID: "user1", wantMigrate: true},
		{name: "https ignores an anonymous tossed session", secure: true, cookies: "wdy-ora-access=anon", wantID: ""},
		{name: "https ignores ambiguous sessions", secure: true, cookies: "wdy-ora-access=user-evil; wdy-ora-access=user1", wantID: ""},
		{name: "https without cookie", secure: true, wantID: ""},
	} {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		if tc.cookies != "" {
			r.Header.Set("Cookie", tc.cookies)
		}

		id, migrate := sessionIDFromCookies(r, tc.secure, loggedIn)
		if id != tc.wantID || migrate != tc.wantMigrate {
			t.Errorf("%s: got %q %v, want %q %v", tc.name, id, migrate, tc.wantID, tc.wantMigrate)
		}
	}
}

// The scope id of the wire grants access to a window, so the request log shows only its fingerprint.
func TestRedactURLHidesTheScopeID(t *testing.T) {
	sid := strings.Repeat("a", 40)
	u, _ := url.Parse("/wire?_sid=" + sid + "&x=1")

	got := redactURL(u)
	if strings.Contains(got, sid) || !strings.Contains(got, "x=1") || !strings.Contains(got, "_sid=%23") {
		t.Fatalf("unexpected redacted url %q", got)
	}

	plain, _ := url.Parse("/api/doc?x=1")
	if redactURL(plain) != "/api/doc?x=1" {
		t.Fatal("an url without scope id must stay unchanged")
	}
}
