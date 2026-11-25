// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
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

const linkPassword = "Sup3r-Geheim!2026"

// linkServer is an application served through a real HTTP server like [nagotest.Serve], which additionally exposes
// the core application to look up the scope of a window and the use cases to log sessions in and out.
type linkServer struct {
	t        *testing.T
	base     string
	app      *core.Application
	users    user.UseCases
	sessions session.UseCases
	// refusal is the body of the answer to an unknown link, which every refused request must repeat.
	refusal []byte
}

func serveLinks(t *testing.T, views func(cfg *application.Configurator)) *linkServer {
	t.Helper()

	cfg := application.NewConfigurator()
	cfg.SetDataDir(t.TempDir())
	cfg.SetApplicationID("de.worldiety.linkownertest")
	std.Must(cfg.SessionManagement())
	views(cfg)

	handler, app, err := cfg.NewHTTPHandler()
	if err != nil {
		t.Fatalf("cannot create application: %v", err)
	}

	srv := httptest.NewServer(handler)
	t.Cleanup(func() {
		srv.CloseClientConnections()
		srv.Close()
		app.Destroy()
	})

	return &linkServer{
		t:        t,
		base:     srv.URL,
		app:      app,
		users:    std.Must(cfg.UserManagement()).UseCases,
		sessions: std.Must(cfg.SessionManagement()).UseCases,
	}
}

// person creates a verified user.
func (s *linkServer) person(mail string) user.ID {
	s.t.Helper()

	usr, err := s.users.Create(user.SU(), user.ShortRegistrationUser{Firstname: "A", Lastname: "B", Email: user.Email(mail), Password: linkPassword, PasswordRepeated: linkPassword, Verified: true})
	if err != nil {
		s.t.Fatal(err)
	}

	return usr.ID
}

func (s *linkServer) login(sid session.ID, usr user.ID) {
	s.t.Helper()

	if err := s.sessions.LoginUser(sid, usr); err != nil {
		s.t.Fatal(err)
	}
}

func (s *linkServer) logout(sid session.ID) {
	s.t.Helper()

	if _, err := s.sessions.Logout(sid); err != nil {
		s.t.Fatal(err)
	}
}

// follow requests the link like a browser of the given session over plain http. An empty session sends no cookie.
// Redirects are not followed, thus the status is the answer of the link itself.
func (s *linkServer) follow(link string, sid session.ID) (status int, body []byte, err error) {
	s.t.Helper()

	ref, err := url.Parse(link)
	if err != nil {
		s.t.Fatalf("invalid link %q: %v", link, err)
	}

	base, _ := url.Parse(s.base)
	req, err := http.NewRequest(http.MethodGet, base.ResolveReference(ref).String(), nil)
	if err != nil {
		s.t.Fatal(err)
	}

	if sid != "" {
		req.AddCookie(&http.Cookie{Name: "wdy-ora-access", Value: string(sid)})
	}

	client := http.Client{
		Timeout:       10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	return resp.StatusCode, body, err
}

// expect requests the link and expects the exact bytes, or a 404 like for an unknown link if want is nil.
func (s *linkServer) expect(row string, link string, sid session.ID, want []byte) {
	s.t.Helper()

	status, body, err := s.follow(link, sid)
	switch {
	case err != nil:
		s.t.Errorf("%s: %v", row, err)
	case want != nil && (status != http.StatusOK || !bytes.Equal(body, want)):
		s.t.Errorf("%s: want 200 with %q, got %d with %q", row, want, status, body)
	case want == nil && status != http.StatusNotFound:
		s.t.Errorf("%s: want 404, got %d with %q", row, status, body)
	case want == nil && s.refusal != nil && !bytes.Equal(body, s.refusal):
		s.t.Errorf("%s: a refusal must look like an unknown link %q, got %q", row, s.refusal, body)
	}
}

// learnRefusal remembers the answer to an unknown link, which every refusal must repeat.
func (s *linkServer) learnRefusal(unknown string) {
	s.t.Helper()

	status, body, err := s.follow(unknown, "")
	if err != nil || status != http.StatusNotFound {
		s.t.Fatalf("an unknown link must answer 404, got %d with %q: %v", status, body, err)
	}

	s.refusal = body
}

// lastResource returns the link of the last download, which the window received.
func lastResource(t *testing.T, w *nagotest.Window) string {
	t.Helper()

	res := w.Resources()
	if len(res) == 0 {
		t.Fatal("the window received no download")
	}

	return string(res[len(res)-1].URI)
}

// forged returns the link with another token of the same length.
func forged(t *testing.T, link string) string {
	t.Helper()

	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("invalid link %q: %v", link, err)
	}

	token := u.Query().Get("token")
	other := strings.Repeat("A", max(len(token), 32))
	if other == token {
		other = strings.Repeat("B", len(other))
	}

	u.RawQuery = url.Values{"token": {other}}.Encode()
	return u.String()
}

// expectForm checks that the link has the form path?token=<token> and names none of the secrets.
func expectForm(t *testing.T, link, path string, secrets map[string]string) {
	t.Helper()

	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("invalid link %q: %v", link, err)
	}

	if u.Path != path {
		t.Errorf("link %q: want the path %s", link, path)
	}

	if q := u.Query(); len(q) != 1 || len(q["token"]) != 1 || q.Get("token") == "" {
		t.Errorf("link %q: want a token as the only query value", link)
	}

	for name, secret := range secrets {
		if strings.Contains(link, secret) {
			t.Errorf("link %q names the %s %q", link, name, secret)
		}
	}
}

func who(wnd core.Window) string {
	if s := wnd.Subject(); s != nil && s.Valid() {
		return s.Email()
	}

	return "anonymous"
}

func report(who string) []byte {
	return []byte("report of " + who + "\x00\xff")
}

// exportPage shows the person of the window and offers a report of that person as a download.
func exportPage(wnd core.Window) core.View {
	return ui.VStack(
		ui.Text("user: "+who(wnd)),
		ui.PrimaryButton(func() {
			wnd.ExportFiles(core.ExportFileBytes("report.txt", report(who(wnd))))
		}).Title("export"),
	)
}

// A download link is delivered only to the session, whose window created it, and only while that session is logged
// in as the person of the window at that time. A link created without login belongs to the session alone. Every
// refusal looks like an unknown link.
func TestDownloadLinkFollowsOnlyItsOwner(t *testing.T) {
	s := serveLinks(t, func(cfg *application.Configurator) {
		cfg.RootView(".", exportPage)
	})

	p := s.person("p@example.com")
	q := s.person("q@example.com")
	sidP := session.ID(strings.Repeat("p", 40))
	sidQ := session.ID(strings.Repeat("q", 40))
	sidAnon := session.ID(strings.Repeat("n", 40))
	sidOther := session.ID(strings.Repeat("o", 40))
	s.login(sidP, p)
	s.login(sidQ, q)

	wp := nagotest.Dial(t, s.base, ".", nagotest.Session(string(sidP)))
	wp.Find(nagotest.Text("user: p@example.com"))
	wp.Click(wp.Find(nagotest.Text("export")))
	linkP := lastResource(t, wp)

	wa := nagotest.Dial(t, s.base, ".", nagotest.Session(string(sidAnon)))
	wa.Find(nagotest.Text("user: anonymous"))
	wa.Click(wa.Find(nagotest.Text("export")))
	linkAnon := lastResource(t, wa)

	s.learnRefusal(forged(t, linkP))

	// the rows run in order, each applies its change of the sessions before its request
	rows := []struct {
		name   string
		before func()
		link   string
		cookie session.ID
		want   []byte // nil expects 404
	}{
		{name: "creating session, logged in as P", link: linkP, cookie: sidP, want: report("p@example.com")},
		{name: "another session, logged in as Q", link: linkP, cookie: sidQ},
		{name: "no cookie", link: linkP},
		{name: "creating session after P logged out", before: func() { s.logout(sidP) }, link: linkP, cookie: sidP},
		{name: "creating session, now logged in as Q", before: func() { s.login(sidP, q) }, link: linkP, cookie: sidP},
		{name: "creating session, logged in as P again", before: func() { s.logout(sidP); s.login(sidP, p) }, link: linkP, cookie: sidP, want: report("p@example.com")},
		{name: "without login: creating session", link: linkAnon, cookie: sidAnon, want: report("anonymous")},
		{name: "without login: another session", link: linkAnon, cookie: sidOther},
		{name: "without login: no cookie", link: linkAnon},
		{name: "without login: creating session, since logged in as Q", before: func() { s.login(sidAnon, q) }, link: linkAnon, cookie: sidAnon, want: report("anonymous")},
		{name: "unknown token", link: forged(t, linkP), cookie: sidP},
	}

	for _, row := range rows {
		if row.before != nil {
			row.before()
		}

		s.expect(row.name, row.link, row.cookie, row.want)
	}

	// proves LINK-DOWNLOAD-OWNER
}

// A download link names neither the window nor the export, and a later export of the same ID replaces it.
func TestDownloadLinkIsOpaque(t *testing.T) {
	const exportID = "export-id-4711"
	first, second := []byte("first version"), []byte("second version")

	s := serveLinks(t, func(cfg *application.Configurator) {
		cfg.RootView(".", func(wnd core.Window) core.View {
			export := func(content []byte) func() {
				return func() {
					wnd.ExportFiles(core.ExportFilesOptions{ID: exportID, Files: []core.File{core.MemFile{Filename: "data.txt", Bytes: content}}})
				}
			}

			return ui.VStack(
				ui.PrimaryButton(export(first)).Title("export first"),
				ui.PrimaryButton(export(second)).Title("export second"),
			)
		})
	})

	sid := session.ID(strings.Repeat("s", 40))
	w := nagotest.Dial(t, s.base, ".", nagotest.Session(string(sid)))
	w.Click(w.Find(nagotest.Text("export first")))
	firstLink := lastResource(t, w)
	secrets := map[string]string{"window id": string(w.ScopeID()), "export id": exportID}
	expectForm(t, firstLink, "/api/ora/v1/download", secrets)

	var secondLink string
	rows := []struct {
		name   string
		before func()
		link   func() string
		want   []byte // nil expects 404
	}{
		{name: "first link before a second export", link: func() string { return firstLink }, want: first},
		{name: "first link after a second export of the same ID", before: func() {
			w.Click(w.Find(nagotest.Text("export second")))
			secondLink = lastResource(t, w)
		}, link: func() string { return firstLink }},
		{name: "second link", link: func() string { return secondLink }, want: second},
	}

	for _, row := range rows {
		if row.before != nil {
			row.before()
		}

		s.expect(row.name, row.link(), sid, row.want)
	}

	expectForm(t, secondLink, "/api/ora/v1/download", secrets)
	if secondLink == firstLink {
		t.Errorf("the second export repeated the link %q", firstLink)
	}

	// proves LINK-DOWNLOAD-OPAQUE
}

type sharer interface {
	AsURI(open func() (io.Reader, error)) (core.URI, error)
}

// A link of AsURI follows the owner rule of a download and lives until the window ends, across a change of page.
func TestShareLinkFollowsOnlyItsOwner(t *testing.T) {
	content := []byte("shared stream\x00\xff")

	s := serveLinks(t, func(cfg *application.Configurator) {
		cfg.RootView(".", func(wnd core.Window) core.View {
			shared := core.AutoState[string](wnd)
			return ui.VStack(
				ui.Text("user: "+who(wnd)),
				ui.PrimaryButton(func() {
					sh, ok := wnd.(sharer)
					if !ok {
						shared.Set("share failed: the window has no AsURI")
						return
					}

					uri, err := sh.AsURI(func() (io.Reader, error) { return bytes.NewReader(content), nil })
					if err != nil {
						shared.Set("share failed: " + err.Error())
						return
					}

					shared.Set("shared: " + string(uri))
				}).Title("share"),
				ui.Text(shared.Get()),
				ui.PrimaryButton(func() { wnd.Navigation().ForwardTo("next", nil) }).Title("go on"),
			)
		})

		cfg.RootView("next", func(wnd core.Window) core.View {
			return ui.Text("next page")
		})
	})

	p := s.person("p@example.com")
	q := s.person("q@example.com")
	sidP := session.ID(strings.Repeat("p", 40))
	sidQ := session.ID(strings.Repeat("q", 40))
	s.login(sidP, p)
	s.login(sidQ, q)

	w := nagotest.Dial(t, s.base, ".", nagotest.Session(string(sidP)))
	w.Find(nagotest.Text("user: p@example.com"))
	w.Click(w.Find(nagotest.Text("share")))
	w.FindAll(nagotest.TextContains("share failed")).None()
	link := strings.TrimPrefix(w.Find(nagotest.TextContains("shared: ")).Text(), "shared: ")
	expectForm(t, link, "/api/ora/v1/share", map[string]string{"window id": string(w.ScopeID())})

	s.learnRefusal(forged(t, link))

	// the rows run in order, each applies its change before its request
	rows := []struct {
		name   string
		before func()
		cookie session.ID
		link   string
		want   []byte // nil expects 404
	}{
		{name: "creating session, logged in as P", cookie: sidP, link: link, want: content},
		{name: "another session, logged in as Q", cookie: sidQ, link: link},
		{name: "no cookie", link: link},
		{name: "creating session after a change of page", before: func() {
			w.Click(w.Find(nagotest.Text("go on")))
			w.Find(nagotest.Text("next page"))
		}, cookie: sidP, link: link, want: content},
		{name: "creating session after P logged out", before: func() { s.logout(sidP) }, cookie: sidP, link: link},
		{name: "creating session, logged in as P again", before: func() { s.login(sidP, p) }, cookie: sidP, link: link, want: content},
		{name: "creating session after the window ended", before: func() {
			if !s.app.DestroyScope(w.ScopeID()) {
				t.Fatal("the window has no scope to end")
			}
		}, cookie: sidP, link: link},
		{name: "unknown token", cookie: sidP, link: forged(t, link)},
	}

	for _, row := range rows {
		if row.before != nil {
			row.before()
		}

		s.expect(row.name, row.link, row.cookie, row.want)
	}

	// proves LINK-SHARE-OWNER
}

// Following a download link, allowed or not, does not keep its window alive, unlike a message of the window.
func TestLinkRequestDoesNotExtendTheWindow(t *testing.T) {
	content := []byte("kept")

	s := serveLinks(t, func(cfg *application.Configurator) {
		cfg.RootView(".", func(wnd core.Window) core.View {
			pings := core.AutoState[int](wnd)
			return ui.VStack(
				ui.PrimaryButton(func() { wnd.ExportFiles(core.ExportFileBytes("kept.txt", content)) }).Title("export"),
				ui.PrimaryButton(func() { pings.Set(pings.Get() + 1) }).Title("ping"),
				ui.Text(fmt.Sprintf("pings: %d", pings.Get())),
			)
		})
	})

	sid := session.ID(strings.Repeat("k", 40))
	w := nagotest.Dial(t, s.base, ".", nagotest.Session(string(sid)))
	w.Click(w.Find(nagotest.Text("export")))
	link := lastResource(t, w)

	scope, ok := s.app.Scope(w.ScopeID())
	if !ok {
		t.Fatal("the window has no scope")
	}

	// the rows run in order: the message needs the connection, the requests run after it is closed, thus no keepalive
	// or message of the window can move the end of life in between
	rows := []struct {
		name    string
		before  func()
		act     func(row string)
		extends bool
	}{
		{name: "message of the window", act: func(string) {
			w.Click(w.Find(nagotest.Text("ping")))
			w.Find(nagotest.Text("pings: 1"))
		}, extends: true},
		{name: "allowed request", before: w.Close, act: func(row string) { s.expect(row, link, sid, content) }},
		{name: "refused request", act: func(row string) { s.expect(row, link, session.ID(strings.Repeat("o", 40)), nil) }},
		{name: "unknown token", act: func(row string) { s.expect(row, forged(t, link), sid, nil) }},
	}

	for _, row := range rows {
		if row.before != nil {
			row.before()
		}

		eol := scope.EOL()
		row.act(row.name)
		after := scope.EOL()

		if extended := after.After(eol); extended != row.extends || (!row.extends && !after.Equal(eol)) {
			t.Errorf("%s: end of life moved from %v to %v, want extended %v", row.name, eol, after, row.extends)
		}
	}

	// proves LINK-NO-EXTEND
}

// slowSource yields size bytes in chunks and pauses before each further chunk, like a slow generator of a large
// export. read counts the bytes taken from it.
type slowSource struct {
	size, chunk int
	pause       time.Duration
	pos         int
	read        *atomic.Int64
}

func slowByte(i int) byte {
	return byte(i % 251)
}

func (s *slowSource) Read(p []byte) (int, error) {
	if s.pos >= s.size {
		return 0, io.EOF
	}

	if s.pos > 0 && s.pos%s.chunk == 0 {
		time.Sleep(s.pause)
	}

	end := min(s.size, s.pos-s.pos%s.chunk+s.chunk, s.pos+len(p))
	for i := s.pos; i < end; i++ {
		p[i-s.pos] = slowByte(i)
	}

	n := end - s.pos
	s.pos = end
	s.read.Add(int64(n))
	return n, nil
}

// A transfer which outlasts the limit ends incomplete, the same transfer within the limit delivers everything.
func TestLinkTransferIsCapped(t *testing.T) {
	const chunk, chunks = 8 << 10, 16
	const pause = 10 * time.Millisecond // 15 pauses: the transfer takes at least 150 ms

	want := make([]byte, chunk*chunks)
	for i := range want {
		want[i] = slowByte(i)
	}

	for _, row := range []struct {
		name     string
		limit    time.Duration // zero keeps the default
		complete bool
	}{
		{name: "limit of 50 ms", limit: 50 * time.Millisecond, complete: false},
		{name: "default limit", complete: true},
	} {
		t.Run(row.name, func(t *testing.T) {
			if row.limit > 0 {
				// registered before the server, thus restored after the server completed all requests
				t.Cleanup(application.SetLinkTransferLimit(row.limit))
			}

			var read atomic.Int64
			s := serveLinks(t, func(cfg *application.Configurator) {
				cfg.RootView(".", func(wnd core.Window) core.View {
					return ui.VStack(
						ui.PrimaryButton(func() {
							file := core.NewReaderFile(func() (io.ReadCloser, error) {
								return io.NopCloser(&slowSource{size: len(want), chunk: chunk, pause: pause, read: &read}), nil
							}).SetName("slow.bin")
							wnd.ExportFiles(core.ExportFilesOptions{ID: "slow", Files: []core.File{file}})
						}).Title("export"),
					)
				})
			})

			sid := session.ID(strings.Repeat("t", 40))
			w := nagotest.Dial(t, s.base, ".", nagotest.Session(string(sid)))
			w.Click(w.Find(nagotest.Text("export")))
			status, body, err := s.follow(lastResource(t, w), sid)

			// the premise: the link was followed and the transfer started, thus nothing but the limit can end it
			if read.Load() == 0 {
				t.Fatalf("the transfer did not start: status %d, %v", status, err)
			}

			complete := err == nil && status == http.StatusOK && bytes.Equal(body, want)
			if complete != row.complete {
				t.Errorf("complete %v, want %v: status %d, %d of %d bytes, error %v", complete, row.complete, status, len(body), len(want), err)
			}
		})
	}

	// proves LINK-TRANSFER-CAP
}
