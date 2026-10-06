// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func serveSPA(t *testing.T) string {
	return serveSPADebug(t, false)
}

// serveSPADebug serves the frontend, like production unless debug is set.
func serveSPADebug(t *testing.T, debug bool) string {
	return nagotest.Serve(t, func(cfg *application.Configurator) {
		cfg.Debug(debug)
		cfg.SetApplicationID("de.worldiety.sessioncookietest")
		cfg.Serve(vuejs.Dist())
		cfg.RootView(".", func(wnd core.Window) core.View { return ui.Text("home") })
	})
}

// sessionCookie requests the page and returns the session cookie the server sets, if any.
func sessionCookie(t *testing.T, base, method string, headers map[string]string) *http.Cookie {
	t.Helper()
	req, _ := http.NewRequest(method, base+"/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	for _, c := range res.Cookies() {
		if c.Name == "wdy-ora-access" {
			return c
		}
	}

	return nil
}

// A server behind an https proxy may be used directly by plain http at the same time, e.g. as a fallback in a local
// network. The session cookie is Secure only for https, because a browser drops a Secure cookie of a plain http page.
func secureFollowsTheRequest(t *testing.T, base string) {

	if c := sessionCookie(t, base, http.MethodGet, nil); c == nil || c.Secure {
		t.Fatalf("a plain http page needs a cookie which is not secure, got %+v", c)
	}

	if c := sessionCookie(t, base, http.MethodGet, map[string]string{"X-Forwarded-Proto": "https"}); c == nil || !c.Secure {
		t.Fatalf("a page through an https proxy needs a secure cookie, got %+v", c)
	}
}

// A cross-site navigation or the POST of an identity provider may lack the cookie just because of SameSite. A new
// cookie would replace the session of the user, so it is only issued for same-site and direct requests.
func notIssuedForCrossSiteOrUnsafeRequests(t *testing.T, base string) {

	for name, tc := range map[string]struct {
		method string
		site   string
		want   bool
	}{
		"typed url":              {method: http.MethodGet, site: "none", want: true},
		"same origin asset":      {method: http.MethodGet, site: "same-origin", want: true},
		"old browser":            {method: http.MethodGet, want: true},
		"link from another site": {method: http.MethodGet, site: "cross-site", want: false},
		"post of an idp":         {method: http.MethodPost, site: "cross-site", want: false},
		"same origin post":       {method: http.MethodPost, site: "same-origin", want: false},
	} {
		headers := map[string]string{}
		if tc.site != "" {
			headers["Sec-Fetch-Site"] = tc.site
		}

		if got := sessionCookie(t, base, tc.method, headers) != nil; got != tc.want {
			t.Errorf("%s: cookie issued %v, want %v", name, got, tc.want)
		}
	}
}

// The wire issues the cookie, if the page could not get one, e.g. because its assets came from the browser cache.
func wireIssuesMissingSessionCookie(t *testing.T, base string) {

	u, _ := url.Parse(base)
	u.Scheme = "ws"
	u.Path = "/wire"
	u.RawQuery = url.Values{"_sid": {strings.Repeat("s", 40)}}.Encode()

	conn, res, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var issued *http.Cookie
	for _, c := range res.Cookies() {
		if c.Name == "wdy-ora-access" {
			issued = c
		}
	}

	if issued == nil || issued.Value == "" || !issued.HttpOnly || issued.Secure {
		t.Fatalf("expected the handshake to issue an http-only cookie for plain http, got %+v", issued)
	}

	// with a cookie, the handshake issues nothing
	header := http.Header{}
	header.Set("Cookie", (&http.Cookie{Name: "wdy-ora-access", Value: issued.Value}).String())
	conn2, res2, err := websocket.DefaultDialer.Dial(strings.Replace(u.String(), strings.Repeat("s", 40), strings.Repeat("t", 40), 1), header)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()

	if len(res2.Cookies()) != 0 {
		t.Fatalf("the handshake must keep an existing cookie, got %v", res2.Cookies())
	}
}

// The former restore endpoint set the session cookie to any id encrypted with the master key. It is gone, and an
// http flow does not hand out the session id anymore.
func restoreIsGone(t *testing.T, base string) {

	res, err := http.Post(base+"/api/nago/v1/session/restore", "text/plain", strings.NewReader("00ff"))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	for _, c := range res.Cookies() {
		if c.Name == "wdy-ora-access" {
			t.Fatalf("the restore endpoint must not set a cookie, got %+v", c)
		}
	}

	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.httpflowtest")
		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.PrimaryButton(func() {
				core.HTTPFlow(wnd.Navigation(), "https://idp.example.com/authorize", "https://app.example.com/callback", "callback")
			}).Title("login")
		})
	})

	w := app.Open(t, nil, ".")
	w.Click(w.Find(nagotest.Text("login")))

	var flow *proto.OpenHttpFlow
	for _, evt := range w.Events() {
		if f, ok := evt.(*proto.OpenHttpFlow); ok {
			flow = f
		}
	}

	if flow == nil || flow.Url != "https://idp.example.com/authorize" || flow.Session != "" {
		t.Fatalf("expected a flow without session, got %+v", flow)
	}
}

// TestSessionCookie shares one server, because serving the frontend compresses all its assets at start.
func TestSessionCookie(t *testing.T) {
	base := serveSPA(t)
	t.Run("secure follows the request", func(t *testing.T) { secureFollowsTheRequest(t, base) })
	t.Run("not issued for cross-site or unsafe requests", func(t *testing.T) { notIssuedForCrossSiteOrUnsafeRequests(t, base) })
	t.Run("wire issues a missing cookie", func(t *testing.T) { wireIssuesMissingSessionCookie(t, base) })
	t.Run("restore is gone", func(t *testing.T) { restoreIsGone(t, base) })
	t.Run("origin of the wire", func(t *testing.T) { originOfTheWire(t, base) })
}

// wireOrigin dials the wire with the given Origin and reports whether the server accepted the handshake.
func wireOrigin(t *testing.T, base string, headers map[string]string) bool {
	t.Helper()
	u, _ := url.Parse(base)
	u.Scheme = "ws"
	u.Path = "/wire"
	u.RawQuery = url.Values{"_sid": {strings.Repeat("o", 40)}}.Encode()

	header := http.Header{}
	for k, v := range headers {
		header.Set(k, v)
	}

	conn, res, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		if res == nil || res.StatusCode != http.StatusForbidden {
			t.Fatalf("expected a refused handshake, got %v", err)
		}
		return false
	}

	_ = conn.Close()
	return true
}

// A page of another application on the same parent domain must not open the wire with the cookie of the user.
func originOfTheWire(t *testing.T, base string) {
	host := strings.TrimPrefix(base, "http://")
	for name, tc := range map[string]struct {
		headers map[string]string
		want    bool
	}{
		"own page":               {headers: map[string]string{"Origin": base}, want: true},
		"native client":          {want: true},
		"sibling subdomain":      {headers: map[string]string{"Origin": "https://evil.apps.example.com"}, want: false},
		"same host, other port":  {headers: map[string]string{"Origin": "http://" + strings.Split(host, ":")[0] + ":1"}, want: false},
		"through a proxy":        {headers: map[string]string{"Origin": "https://app.apps.example.com", "X-Forwarded-Host": "app.apps.example.com"}, want: true},
		"proxy with rfc 7239":    {headers: map[string]string{"Origin": "https://app.apps.example.com", "Forwarded": `host="app.apps.example.com";proto=https`}, want: true},
		"proxy for another host": {headers: map[string]string{"Origin": "https://evil.apps.example.com", "X-Forwarded-Host": "app.apps.example.com"}, want: false},
	} {
		if got := wireOrigin(t, base, tc.headers); got != tc.want {
			t.Errorf("%s: accepted %v, want %v", name, got, tc.want)
		}
	}
}

// The development server of the frontend may run on another port of the local machine, but only in debug mode.
func TestWireOriginOfTheDevelopmentServer(t *testing.T) {
	base := serveSPADebug(t, true)
	host := strings.Split(strings.TrimPrefix(base, "http://"), ":")[0]
	if !wireOrigin(t, base, map[string]string{"Origin": "http://" + host + ":8090"}) {
		t.Fatal("a local development server must be accepted in debug mode")
	}

	if wireOrigin(t, base, map[string]string{"Origin": "https://evil.apps.example.com"}) {
		t.Fatal("a foreign origin must be refused also in debug mode")
	}
}
