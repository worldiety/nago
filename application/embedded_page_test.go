// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"cmp"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gorilla/websocket"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// shell is delivered for the path of every page, like index.html of the frontend.
var shell = fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>shell</title>")}}

// embeddedPage shows the person of the window and a counter, whose state only the window keeps, and offers a report
// of the person as a download.
func embeddedPage(wnd core.Window) core.View {
	count := core.AutoState[int](wnd)
	return ui.VStack(
		ui.Text("user: "+who(wnd)),
		ui.Text(fmt.Sprintf("count: %d", count.Get())),
		ui.PrimaryButton(func() { count.Set(count.Get() + 1) }).Title("increment"),
		ui.PrimaryButton(func() {
			wnd.ExportFiles(core.ExportFileBytes("report.txt", report(who(wnd))))
		}).Title("export"),
	)
}

// serveEmbedded serves the shell and the same page under ., e/* and x/page, whereof the given paths are embeddable.
func serveEmbedded(t *testing.T, embeddable ...core.NavigationPath) *linkServer {
	t.Helper()

	return serveLinks(t, func(cfg *application.Configurator) {
		cfg.Serve(shell)
		cfg.EmbeddablePages(embeddable...)
		cfg.RootView(".", embeddedPage)
		cfg.RootView("e/*", embeddedPage)
		cfg.RootView("x/page", embeddedPage)
	})
}

// navigation is the request of a browser for a page.
type navigation struct {
	method      string // GET if empty
	path        string
	site, dest  string // the headers Sec-Fetch-Site and Sec-Fetch-Dest; empty sends none
	https       bool   // through an https proxy
	host        string // the Host header; empty keeps the local address of the server
	cookies     []*http.Cookie
	ifNoneMatch string
}

// inForeignFrame is the navigation of a frame of another site to the path.
func inForeignFrame(path string) navigation {
	return navigation{path: path, site: "cross-site", dest: "iframe"}
}

// direct is the navigation of the address bar to the path.
func direct(path string) navigation {
	return navigation{path: path, site: "none", dest: "document"}
}

func (n navigation) overHTTPS() navigation {
	n.https = true
	return n
}

func (n navigation) withCookies(cookies ...*http.Cookie) navigation {
	n.cookies = cookies
	return n
}

// navigate sends the navigation and returns the answer without its body. Redirects are not followed.
func (s *linkServer) navigate(t *testing.T, n navigation) *http.Response {
	t.Helper()

	req, err := http.NewRequest(cmp.Or(n.method, http.MethodGet), s.base+n.path, nil)
	if err != nil {
		t.Fatal(err)
	}

	if n.host != "" {
		req.Host = n.host
	}

	for k, v := range map[string]string{"Sec-Fetch-Site": n.site, "Sec-Fetch-Dest": n.dest, "If-None-Match": n.ifNoneMatch} {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	if n.https {
		req.Header.Set("X-Forwarded-Proto", "https")
	}

	for _, c := range n.cookies {
		req.AddCookie(c)
	}

	client := http.Client{
		Timeout:       nagotest.SettleTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp
}

// embeddedCookieOf returns the embedded session, which the navigation of a frame of another site to the path issues.
func (s *linkServer) embeddedCookieOf(t *testing.T, path string) session.ID {
	t.Helper()

	for _, c := range s.navigate(t, inForeignFrame(path)).Cookies() {
		if c.Name == "wdy-ora-embedded" {
			return session.ID(c.Value)
		}
	}

	t.Fatalf("the frame of another site got no embedded cookie for %s", path)
	return ""
}

var (
	embeddedCookieNames = []string{"wdy-ora-embedded", "__Host-wdy-ora-embedded"}
	plainCookieNames    = []string{"wdy-ora-access", "__Host-wdy-ora-access"}
)

// cookieWant is the answer to a navigation.
type cookieWant struct {
	status   int    // 200 if 0
	embedded string // the name of the embedded cookie; empty expects none
	keeps    string // the value, which the embedded cookie keeps; empty expects a new embedded id
	plain    string // the name of the plain cookie; empty expects none
	noCache  bool   // expects Cache-Control: no-cache, otherwise it is not asserted
}

func expectCookies(t *testing.T, row string, n navigation, resp *http.Response, want cookieWant) {
	t.Helper()

	if status := cmp.Or(want.status, http.StatusOK); resp.StatusCode != status {
		t.Errorf("%s: status %d, want %d", row, resp.StatusCode, status)
	}

	if want.noCache && resp.Header.Get("Cache-Control") != "no-cache" {
		t.Errorf("%s: Cache-Control %q, want no-cache", row, resp.Header.Get("Cache-Control"))
	}

	var embedded, plain []*http.Cookie
	for _, c := range resp.Cookies() {
		switch {
		case slices.Contains(embeddedCookieNames, c.Name):
			embedded = append(embedded, c)
		case slices.Contains(plainCookieNames, c.Name):
			plain = append(plain, c)
		default:
			t.Errorf("%s: unexpected cookie %s", row, c)
		}
	}

	switch {
	case want.embedded == "" && len(embedded) > 0:
		t.Errorf("%s: want no embedded cookie, got %v", row, embedded)
	case want.embedded != "" && (len(embedded) != 1 || embedded[0].Name != want.embedded):
		t.Errorf("%s: want one embedded cookie %s, got %v", row, want.embedded, embedded)
	case want.embedded != "":
		c := embedded[0]
		if !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteNoneMode || !c.Partitioned || c.Path != "/" || c.Domain != "" || !c.Expires.After(time.Now()) {
			t.Errorf("%s: want a Secure, HttpOnly, SameSite=None, Partitioned cookie for / without domain, got %s", row, c)
		}

		if !session.ID(c.Value).Embedded() {
			t.Errorf("%s: the embedded cookie carries no embedded session: %q", row, c.Value)
		}

		sent := slices.ContainsFunc(n.cookies, func(s *http.Cookie) bool { return s.Value == c.Value })
		switch {
		case want.keeps != "" && c.Value != want.keeps:
			t.Errorf("%s: the embedded cookie carries %q, want it to keep %q", row, c.Value, want.keeps)
		case want.keeps == "" && sent:
			t.Errorf("%s: the embedded cookie carries the sent value %q, want a new one", row, c.Value)
		}
	}

	switch {
	case want.plain == "" && len(plain) > 0:
		t.Errorf("%s: want no plain cookie, got %v", row, plain)
	case want.plain != "" && (len(plain) != 1 || plain[0].Name != want.plain):
		t.Errorf("%s: want one plain cookie %s, got %v", row, want.plain, plain)
	case want.plain != "" && plain[0].SameSite != http.SameSiteLaxMode:
		t.Errorf("%s: want the plain cookie with SameSite=Lax, got %s", row, plain[0])
	}
}

// The navigation of a frame of another site to an embeddable page gets the embedded cookie, which is SameSite=None,
// Secure and Partitioned, over https with the prefix __Host-, over http://localhost without. Its answer is never
// taken from the cache without asking and issues the cookie also when it only confirms the cached page. The same page
// direct, by a link of another site or in a frame of the same site, another page in a frame of another site, plain
// http of another host and NO_SSL=true get no embedded cookie, and the navigation of a frame never touches the plain
// cookie.
func TestEmbeddedCookieOnlyInAFrameOfAnotherSite(t *testing.T) {
	s := serveEmbedded(t, "e/*")
	const page = "/e/forms/contact"
	etag := s.navigate(t, direct(page)).Header.Get("Etag")
	if etag == "" {
		t.Fatal("the page has no ETag, thus no row can revalidate it")
	}

	embedded := session.NewEmbeddedID()
	loggedIn := session.ID(strings.Repeat("p", 40))
	s.login(loggedIn, s.person("p@example.com"))

	revalidate := func(n navigation) navigation {
		n.ifNoneMatch = etag
		return n
	}

	otherHost := func(n navigation) navigation {
		n.host = "app.example.com"
		return n
	}

	post := inForeignFrame(page).overHTTPS()
	post.method = http.MethodPost

	sameSite := inForeignFrame(page).overHTTPS()
	sameSite.site = "same-site"

	link := direct(page).overHTTPS()
	link.site = "cross-site"

	frameset := inForeignFrame(page).overHTTPS()
	frameset.dest = "frame"

	rows := []struct {
		name string
		nav  navigation
		want cookieWant
	}{
		{name: "frame of another site, https", nav: inForeignFrame(page).overHTTPS(), want: cookieWant{embedded: "__Host-wdy-ora-embedded", noCache: true}},
		{name: "frameset of another site, https", nav: frameset, want: cookieWant{embedded: "__Host-wdy-ora-embedded", noCache: true}},
		{name: "direct, https", nav: direct(page).overHTTPS(), want: cookieWant{plain: "__Host-wdy-ora-access"}},
		{name: "link of another site, https", nav: link},
		{name: "frame of the same site, https", nav: sameSite, want: cookieWant{plain: "__Host-wdy-ora-access"}},
		{name: "frame of another site, another page, https", nav: inForeignFrame("/x/page").overHTTPS()},
		{name: "frame of another site, the root page, https", nav: inForeignFrame("/").overHTTPS()},
		{name: "form of another site into a frame, https", nav: post, want: cookieWant{status: http.StatusMethodNotAllowed}},
		{name: "frame of another site, localhost", nav: inForeignFrame(page), want: cookieWant{embedded: "wdy-ora-embedded", noCache: true}},
		{name: "frame of another site, plain http of another host", nav: otherHost(inForeignFrame(page)), want: cookieWant{noCache: true}},
		{name: "direct, plain http of another host", nav: otherHost(direct(page)), want: cookieWant{plain: "wdy-ora-access"}},
		{name: "frame of another site, https, with its embedded cookie", nav: inForeignFrame(page).overHTTPS().withCookies(&http.Cookie{Name: "__Host-wdy-ora-embedded", Value: string(embedded)}), want: cookieWant{embedded: "__Host-wdy-ora-embedded", keeps: string(embedded), noCache: true}},
		{name: "frame of another site, https, with an embedded cookie of a plain session", nav: inForeignFrame(page).overHTTPS().withCookies(&http.Cookie{Name: "__Host-wdy-ora-embedded", Value: string(loggedIn)}), want: cookieWant{embedded: "__Host-wdy-ora-embedded", noCache: true}},
		{name: "frame of another site, https, revalidation", nav: revalidate(inForeignFrame(page).overHTTPS()), want: cookieWant{status: http.StatusNotModified, embedded: "__Host-wdy-ora-embedded", noCache: true}},
		{name: "direct, https, revalidation", nav: revalidate(direct(page).overHTTPS()), want: cookieWant{status: http.StatusNotModified, plain: "__Host-wdy-ora-access"}},
		{name: "frame of another site, https, with the plain cookie of a logged-in session", nav: inForeignFrame(page).overHTTPS().withCookies(&http.Cookie{Name: "__Host-wdy-ora-access", Value: string(loggedIn)}), want: cookieWant{embedded: "__Host-wdy-ora-embedded", noCache: true}},
		{name: "direct, https, with the plain cookie of a logged-in session", nav: direct(page).overHTTPS().withCookies(&http.Cookie{Name: "__Host-wdy-ora-access", Value: string(loggedIn)})},
	}

	for _, row := range rows {
		expectCookies(t, row.name, row.nav, s.navigate(t, row.nav), row.want)
	}

	if s.sessions.FindUserSessionByID(loggedIn).User().IsNone() {
		t.Error("the logged-in session lost its login")
	}

	// proves NAGO-EMBED-COOKIE
}

// Only the configuration makes a page embeddable, and NO_SSL decides like for the plain cookie, whether a cookie is
// Secure: NO_SSL=true issues no embedded cookie, NO_SSL=false one with the prefix __Host- also over plain http.
func TestEmbeddedCookieFollowsTheConfiguration(t *testing.T) {
	const page = "/e/forms/contact"
	otherHost := inForeignFrame(page)
	otherHost.host = "app.example.com"

	servers := []struct {
		name       string
		noSSL      string // empty keeps NO_SSL unset
		embeddable []core.NavigationPath
		rows       []struct {
			name string
			nav  navigation
			want cookieWant
		}
	}{
		{name: "without embeddable pages", rows: []struct {
			name string
			nav  navigation
			want cookieWant
		}{
			{name: "frame of another site, https", nav: inForeignFrame(page).overHTTPS()},
			{name: "frame of another site, localhost", nav: inForeignFrame(page)},
			{name: "direct, https", nav: direct(page).overHTTPS(), want: cookieWant{plain: "__Host-wdy-ora-access"}},
		}},
		{name: "embeddable e/*", embeddable: []core.NavigationPath{"e/*"}, rows: []struct {
			name string
			nav  navigation
			want cookieWant
		}{
			{name: "frame of another site, https", nav: inForeignFrame(page).overHTTPS(), want: cookieWant{embedded: "__Host-wdy-ora-embedded", noCache: true}},
			{name: "frame of another site, localhost", nav: inForeignFrame(page), want: cookieWant{embedded: "wdy-ora-embedded", noCache: true}},
		}},
		{name: "embeddable x/page", embeddable: []core.NavigationPath{"x/page"}, rows: []struct {
			name string
			nav  navigation
			want cookieWant
		}{
			{name: "frame of another site, e/, https", nav: inForeignFrame(page).overHTTPS()},
			{name: "frame of another site, x/page, https", nav: inForeignFrame("/x/page").overHTTPS(), want: cookieWant{embedded: "__Host-wdy-ora-embedded", noCache: true}},
		}},
		{name: "NO_SSL=true", noSSL: "true", embeddable: []core.NavigationPath{"e/*"}, rows: []struct {
			name string
			nav  navigation
			want cookieWant
		}{
			{name: "frame of another site, https", nav: inForeignFrame(page).overHTTPS(), want: cookieWant{noCache: true}},
			{name: "frame of another site, localhost", nav: inForeignFrame(page), want: cookieWant{noCache: true}},
			{name: "direct, localhost", nav: direct(page), want: cookieWant{plain: "wdy-ora-access"}},
		}},
		{name: "NO_SSL=false", noSSL: "false", embeddable: []core.NavigationPath{"e/*"}, rows: []struct {
			name string
			nav  navigation
			want cookieWant
		}{
			{name: "frame of another site, plain http of another host", nav: otherHost, want: cookieWant{embedded: "__Host-wdy-ora-embedded", noCache: true}},
		}},
	}

	for _, srv := range servers {
		t.Run(srv.name, func(t *testing.T) {
			if srv.noSSL != "" {
				t.Setenv("NO_SSL", srv.noSSL)
			}

			s := serveEmbedded(t, srv.embeddable...)
			for _, row := range srv.rows {
				expectCookies(t, row.name, row.nav, s.navigate(t, row.nav), row.want)
			}
		})
	}

	// proves NAGO-EMBED-COOKIE
}

// The wire of a page with the embedded cookie issues no cookie, because a browser would drop one with SameSite=Lax in
// the frame. Without cookie, it issues the plain one, also beside an embedded cookie, which carries no embedded
// session.
func TestWireIssuesNoCookieBesideTheEmbeddedOne(t *testing.T) {
	s := serveEmbedded(t, "e/*")
	embedded := s.embeddedCookieOf(t, "/e/forms/contact")

	handshake := func(cookie *http.Cookie) ([]*http.Cookie, error) {
		u, _ := url.Parse(s.base)
		u.Scheme = "ws"
		u.Path = "/wire"
		u.RawQuery = url.Values{"_sid": {string(proto.NewScopeID())}}.Encode()
		header := http.Header{"Origin": {s.base}}
		if cookie != nil {
			header.Set("Cookie", cookie.String())
		}

		conn, resp, err := websocket.DefaultDialer.Dial(u.String(), header)
		if err != nil {
			return nil, err
		}

		_ = conn.Close()
		return resp.Cookies(), nil
	}

	poll := func(cookie *http.Cookie) ([]*http.Cookie, error) {
		b := newBrowser(t, s.base, "")
		header := http.Header{}
		if cookie != nil {
			header.Set("Cookie", cookie.String())
		}

		status, respHeader, body, err := b.post("/wire/poll", url.Values{"_sid": {string(proto.NewScopeID())}}, header, nil)
		if err != nil {
			return nil, err
		}

		if status != http.StatusOK {
			return nil, fmt.Errorf("the poll answered %d: %s", status, body)
		}

		return (&http.Response{Header: respHeader}).Cookies(), nil
	}

	for _, wire := range []struct {
		name    string
		connect func(cookie *http.Cookie) ([]*http.Cookie, error)
	}{
		{name: "websocket", connect: handshake},
		{name: "http", connect: poll},
	} {
		for _, row := range []struct {
			name   string
			cookie *http.Cookie
			want   string // the name of the issued cookie; empty expects none
		}{
			{name: "the embedded cookie", cookie: &http.Cookie{Name: "wdy-ora-embedded", Value: string(embedded)}},
			{name: "no cookie", want: "wdy-ora-access"},
			{name: "an embedded cookie of a plain session", cookie: &http.Cookie{Name: "wdy-ora-embedded", Value: strings.Repeat("p", 40)}, want: "wdy-ora-access"},
		} {
			name := wire.name + ", " + row.name
			cookies, err := wire.connect(row.cookie)
			if err != nil {
				t.Errorf("%s: %v", name, err)
				continue
			}

			var names []string
			for _, c := range cookies {
				names = append(names, c.Name)
			}

			if want := slices.DeleteFunc([]string{row.want}, func(s string) bool { return s == "" }); !slices.Equal(names, want) {
				t.Errorf("%s: issued %v, want %v", name, names, want)
			}
		}
	}

	// proves NAGO-EMBED-COOKIE
}

// A page of the embedded session downloads its files and connects to its window again after an interruption, like
// the same page direct in a session without login: the window keeps its state. The link and the window stay
// unreachable for the frame of another site, which has another embedded session, for a direct session and without
// cookie. A direct, logged-in window of the same instance keeps its person.
func TestEmbeddedSessionDownloadsAndReconnects(t *testing.T) {
	for _, tr := range wires {
		t.Run(tr.name, func(t *testing.T) {
			s := serveEmbedded(t, "e/*")
			loginOwners(s)

			// the navigation of the frame issues the session, which the page then uses for all its requests
			embedded := s.embeddedCookieOf(t, "/e/forms/contact")
			other := s.embeddedCookieOf(t, "/e/forms/contact")
			if other == embedded {
				t.Fatal("two frames without cookie got the same session")
			}

			// the window shows the root page, which a connection of attempt requests: only the frame navigation
			// knows the path, the wire and the download know the session alone
			we := tr.dial(t, s.base, ".", nagotest.Embedded(), nagotest.Session(string(embedded)))
			we.Find(nagotest.Text("user: anonymous"))
			we.Click(we.Find(nagotest.Text("increment")))
			we.Find(nagotest.Text("count: 1"))
			we.Click(we.Find(nagotest.Text("export")))
			link := lastResource(t, we)
			s.learnRefusal(forged(t, link))

			for _, row := range []struct {
				name   string
				cookie session.ID
				want   []byte // nil expects 404
			}{
				{name: "download, the embedded session of the window", cookie: embedded, want: report("anonymous")},
				{name: "download, the frame of another site", cookie: other},
				{name: "download, direct, logged in as P", cookie: sidP},
				{name: "download, direct, without login", cookie: sidAnon},
				{name: "download, no cookie", cookie: ""},
			} {
				s.expect(row.name, link, row.cookie, row.want)
			}

			owned := we.ScopeID()
			for _, row := range []struct {
				name    string
				before  func()
				session session.ID // empty sends no cookie
				want    []string   // the page; nil expects only the request to reload the page
			}{
				{name: "connect, the frame of another site", session: other},
				{name: "connect, direct, logged in as P", session: sidP},
				{name: "connect, direct, without login", session: sidAnon},
				{name: "connect, no cookie", session: ""},
				{name: "connect, the embedded session of the window after an interruption", before: we.Close, session: embedded, want: []string{"user: anonymous", "count: 1"}},
			} {
				if row.before != nil {
					row.before()
				}

				res := tr.attempt(t, s.base, row.session, owned)
				if row.want == nil {
					if err := res.onlyReload(); err != nil {
						t.Errorf("%s: %v", row.name, err)
					}

					continue
				}

				res.expectPage(t, row.name, row.want)
			}

			wp := tr.dial(t, s.base, "x/page", nagotest.Session(string(sidP)))
			wp.Find(nagotest.Text("user: p@example.com"))
		})
	}

	// proves NAGO-EMBED-SESSION
}

// A login with the login page of nago in the embedded session fails and leaves the session without person, while the
// same login in a plain session logs it in.
func TestEmbeddedSessionStaysWithoutLogin(t *testing.T) {
	s := serveEmbedded(t, "e/*")
	s.person("p@example.com")

	for _, row := range []struct {
		name     string
		embedded bool
		want     string
	}{
		{name: "embedded session", embedded: true, want: "user: anonymous"},
		{name: "plain session", embedded: false, want: "user: p@example.com"},
	} {
		sid := session.ID(proto.NewScopeID())
		opts := []nagotest.OpenOption{nagotest.Session(string(sid))}
		if row.embedded {
			sid = session.NewEmbeddedID()
			opts = []nagotest.OpenOption{nagotest.Embedded(), nagotest.Session(string(sid))}
		}

		w := nagotest.Dial(t, s.base, "account/login", opts...)
		w.Type(w.Find(nagotest.Label("E-Mail Adresse")), "p@example.com")
		w.Type(w.Find(nagotest.Label("Kennwort")), linkPassword)
		w.Click(w.Find(nagotest.ID("nago-action-login")))
		w.Close()

		// the next page of the same session
		next := nagotest.Dial(t, s.base, "e/forms/contact", opts...)
		next.Find(nagotest.Text(row.want))
		if got := s.sessions.FindUserSessionByID(sid).User().IsSome(); got == row.embedded {
			t.Errorf("%s: the session has a person %v, want %v", row.name, got, !row.embedded)
		}
	}

	// proves NAGO-EMBED-SESSION
}
