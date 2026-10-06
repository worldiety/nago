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

		cookie := c.newSessionCookie(proxied, "id")
		if cookie.Name != sessionCookieName || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" || cookie.Secure != tc.proxied {
			t.Errorf("NO_SSL=%q: unexpected cookie %+v", tc.noSSL, cookie)
		}
	}
}
