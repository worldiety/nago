// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/core/http/poll"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// windowPage shows the person of the window and a counter, whose state only the window keeps.
func windowPage(wnd core.Window) core.View {
	count := core.AutoState[int](wnd)
	return ui.VStack(
		ui.Text("user: "+who(wnd)),
		ui.Text(fmt.Sprintf("count: %d", count.Get())),
		ui.PrimaryButton(func() { count.Set(count.Get() + 1) }).Title("increment"),
	)
}

type dialer func(t nagotest.TB, baseURL string, path core.NavigationPath, opts ...nagotest.OpenOption) *nagotest.Window

// the request ids of the events, by which a client connects to a window and requests its page, see pageRequests
const (
	ridConfigure proto.RID = 1001
	ridAllocate  proto.RID = 1002
	ridPage      proto.RID = 1003
)

// pageRequests are the events, which the frontend sends after it connected to a window: configure the window, allocate
// the page and, because the backend does not answer the allocation of the page it shows already, request its render,
// see nagotest.Window.Reconnect.
func pageRequests() []proto.NagoEvent {
	return []proto.NagoEvent{
		&proto.ScopeConfigurationChangeRequested{
			RID:            ridConfigure,
			AcceptLanguage: "de",
			WindowInfo:     proto.WindowInfo{Width: 1280, Height: 800, Density: 2, SizeClass: proto.SizeClassXL, ColorScheme: proto.Light},
		},
		&proto.RootViewAllocationRequested{RID: ridAllocate, Locale: "de", Factory: ".", Values: proto.RootViewParameters{}},
		&proto.RootViewRenderingRequested{RID: ridPage},
	}
}

func encode(t *testing.T, evt proto.NagoEvent) []byte {
	t.Helper()

	var buf bytes.Buffer
	if err := proto.Marshal(proto.NewBinaryWriter(&buf), evt); err != nil {
		t.Fatalf("cannot encode %T: %v", evt, err)
	}

	return buf.Bytes()
}

func decode(msg []byte) (proto.NagoEvent, error) {
	obj, err := proto.Unmarshal(proto.NewBinaryReader(bytes.NewBuffer(msg)))
	if err != nil {
		return nil, fmt.Errorf("cannot decode a message: %w", err)
	}

	evt, ok := obj.(proto.NagoEvent)
	if !ok {
		return nil, fmt.Errorf("%T is not a NagoEvent", obj)
	}

	return evt, nil
}

func answers(evt proto.NagoEvent, rid proto.RID) bool {
	src, ok := evt.(interface{ GetRID() proto.RID })
	return ok && src.GetRID() == rid
}

func eventNames(evts []proto.NagoEvent) []string {
	var res []string
	for _, evt := range evts {
		res = append(res, fmt.Sprintf("%T", evt))
	}

	return res
}

// texts returns the texts of all text views of the tree.
func texts(root proto.Component) []string {
	var res []string
	var visit func(v reflect.Value)
	visit = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Interface:
			if !v.IsNil() {
				visit(v.Elem())
			}
		case reflect.Pointer:
			if v.IsNil() {
				return
			}

			if tv, ok := v.Interface().(*proto.TextView); ok {
				res = append(res, string(tv.Value))
			}

			visit(v.Elem())
		case reflect.Struct:
			for i := range v.NumField() {
				if f := v.Field(i); f.CanInterface() {
					visit(f)
				}
			}
		case reflect.Slice, reflect.Array:
			for i := range v.Len() {
				visit(v.Index(i))
			}
		default:
			// scalars, strings and maps contain no text view
		}
	}

	visit(reflect.ValueOf(root))
	return res
}

// wireResult is what a client received, which connected to a window id and requested its page like the frontend.
type wireResult struct {
	events []proto.NagoEvent
	err    error
	// close ends the connection like closing the tab, which otherwise stays open until the test ends
	close func()
}

// onlyReload reports nil, if the client received the request to reload the page and no other event.
func (r wireResult) onlyReload() error {
	if r.err != nil {
		return r.err
	}

	reloads := 0
	var others []proto.NagoEvent
	for _, evt := range r.events {
		if _, ok := evt.(*proto.NavigationReloadRequested); ok {
			reloads++
			continue
		}

		others = append(others, evt)
	}

	switch {
	case len(others) > 0:
		return fmt.Errorf("received %v besides %d requests to reload the page", eventNames(others), reloads)
	case reloads == 0:
		return errors.New("received no request to reload the page")
	}

	return nil
}

// page returns the texts of the page, which the client received last.
func (r wireResult) page() ([]string, error) {
	if r.err != nil {
		return nil, r.err
	}

	var root proto.Component
	for _, evt := range r.events {
		switch evt := evt.(type) {
		case *proto.NavigationReloadRequested:
			return nil, fmt.Errorf("received a request to reload the page: %v", eventNames(r.events))
		case *proto.RootViewInvalidated:
			root = evt.Root
		}
	}

	if root == nil {
		return nil, fmt.Errorf("received no page: %v", eventNames(r.events))
	}

	return texts(root), nil
}

// expectPage checks that the client received a page, which shows all wanted texts.
func (r wireResult) expectPage(t *testing.T, row string, want []string) {
	t.Helper()

	page, err := r.page()
	if err != nil {
		t.Errorf("%s: %v", row, err)
		return
	}

	for _, text := range want {
		if !slices.Contains(page, text) {
			t.Errorf("%s: the page %q does not show %q", row, page, text)
		}
	}
}

// browser sends the cookie of a session like a browser, and keeps a cookie which the backend issues. An empty session
// sends no cookie.
type browser struct {
	base   *url.URL
	jar    http.CookieJar
	client *http.Client
}

func newBrowser(t *testing.T, base string, sid session.ID) *browser {
	t.Helper()

	u, err := url.Parse(base)
	if err != nil {
		t.Fatalf("invalid url %q: %v", base, err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	if sid != "" {
		jar.SetCookies(u, []*http.Cookie{{Name: "wdy-ora-access", Value: string(sid), Path: "/"}})
	}

	return &browser{base: u, jar: jar, client: &http.Client{Jar: jar, Timeout: nagotest.SettleTimeout}}
}

// origin is the header of a page of the application, which a browser sends with each request of the wire.
func (b *browser) origin() http.Header {
	return http.Header{"Origin": {b.base.Scheme + "://" + b.base.Host}}
}

func (b *browser) post(path string, query url.Values, header http.Header, body []byte) (status int, respHeader http.Header, respBody []byte, err error) {
	u := *b.base
	u.Path = path
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return 0, nil, nil, err
	}

	req.Header = b.origin()
	req.Header.Set("Content-Type", "application/octet-stream")
	for k, v := range header {
		req.Header[k] = v
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, respBody, err
}

// wsAttempt connects to the window id through the websocket of the session like the frontend, requests the page and
// reads until the backend answered that request or closed the connection.
func wsAttempt(t *testing.T, base string, sid session.ID, scope proto.ScopeID) wireResult {
	t.Helper()

	b := newBrowser(t, base, sid)
	u := *b.base
	u.Scheme = "ws"
	u.Path = "/wire"
	u.RawQuery = url.Values{"_sid": {string(scope)}}.Encode()

	d := websocket.Dialer{Jar: b.jar, HandshakeTimeout: nagotest.SettleTimeout, EnableCompression: true}
	conn, resp, err := d.Dial(u.String(), b.origin())
	if err != nil {
		if resp != nil {
			return wireResult{err: fmt.Errorf("the handshake answered %d: %w", resp.StatusCode, err), close: func() {}}
		}

		return wireResult{err: fmt.Errorf("the handshake failed: %w", err), close: func() {}}
	}

	// open until the test ends or the row closes it, like a tab, thus closing it cannot change the window in between
	res := wireResult{close: func() { _ = conn.Close() }}
	t.Cleanup(res.close)

	for _, evt := range pageRequests() {
		// like the frontend, the requests follow the handshake at once; a refused connection may be closed already,
		// which the reads below tell
		_ = conn.WriteMessage(websocket.BinaryMessage, encode(t, evt))
	}

	_ = conn.SetReadDeadline(time.Now().Add(nagotest.SettleTimeout))
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			var timeout net.Error
			if errors.As(err, &timeout) && timeout.Timeout() {
				res.err = fmt.Errorf("the backend neither answered the request of the page nor closed the connection within %v, received %v", nagotest.SettleTimeout, eventNames(res.events))
			}

			// otherwise, the backend closed the connection
			return res
		}

		evt, err := decode(msg)
		if err != nil {
			res.err = err
			return res
		}

		res.events = append(res.events, evt)
		if answers(evt, ridPage) {
			return res
		}
	}
}

// wireClient talks to a window through the http wire like the frontend, see nagotest/http.go, and records the
// received events.
type wireClient struct {
	b       *browser
	scope   proto.ScopeID
	token   string // of the channel, which the client polls
	ack     uint64 // the number of the last received message, which the next poll acknowledges
	sendSeq uint64 // the number of the last message, which the backend accepted
	events  []proto.NagoEvent
}

func newWireClient(t *testing.T, base string, sid session.ID, scope proto.ScopeID) *wireClient {
	t.Helper()

	return &wireClient{b: newBrowser(t, base, sid), scope: scope}
}

// as returns a client of another session (none for an empty id), which knows the channel and the numbers of c.
func (c *wireClient) as(t *testing.T, sid session.ID) *wireClient {
	t.Helper()

	return &wireClient{b: newBrowser(t, c.b.base.String(), sid), scope: c.scope, token: c.token, sendSeq: c.sendSeq}
}

// poll fetches the messages after ack and returns the channel, which the answer names. Like the frontend, the client
// takes a new channel over.
func (c *wireClient) poll() (channel string, err error) {
	query := url.Values{"_sid": {string(c.scope)}, "ack": {strconv.FormatUint(c.ack, 10)}}
	if c.token != "" {
		query.Set("_c", c.token)
	}

	status, header, body, err := c.b.post("/wire/poll", query, nil, nil)
	if err != nil {
		return "", fmt.Errorf("poll: %w", err)
	}

	if status != http.StatusOK {
		return "", fmt.Errorf("the poll answered %d: %s", status, bytes.TrimSpace(body))
	}

	channel = header.Get("X-Nago-Channel")
	if channel != "" && channel != c.token {
		c.token, c.ack, c.sendSeq = channel, 0, 0
	}

	first, msgs, err := poll.ReadFrames(bytes.NewReader(body), 1<<30)
	if err != nil {
		return channel, fmt.Errorf("cannot read the frames of the poll: %w", err)
	}

	for i, msg := range msgs {
		seq := first + uint64(i)
		if seq <= c.ack {
			continue
		}

		evt, err := decode(msg)
		if err != nil {
			return channel, err
		}

		c.events = append(c.events, evt)
		c.ack = seq
	}

	return channel, nil
}

// send posts the events as the messages following sendSeq and returns the status of the answer.
func (c *wireClient) send(t *testing.T, evts ...proto.NagoEvent) (int, error) {
	t.Helper()

	msgs := make([][]byte, 0, len(evts))
	for _, evt := range evts {
		msgs = append(msgs, encode(t, evt))
	}

	var body bytes.Buffer
	if err := poll.WriteFrames(&body, c.sendSeq+1, msgs); err != nil {
		t.Fatal(err)
	}

	status, _, _, err := c.b.post("/wire/send", url.Values{"_sid": {string(c.scope)}, "_c": {c.token}}, nil, body.Bytes())
	if err != nil {
		return 0, fmt.Errorf("send: %w", err)
	}

	if status == http.StatusNoContent {
		c.sendSeq += uint64(len(evts))
	}

	return status, nil
}

func (c *wireClient) answered(rid proto.RID) bool {
	return slices.ContainsFunc(c.events, func(evt proto.NagoEvent) bool { return answers(evt, rid) })
}

func (c *wireClient) reloaded() bool {
	return slices.ContainsFunc(c.events, func(evt proto.NagoEvent) bool {
		_, ok := evt.(*proto.NavigationReloadRequested)
		return ok
	})
}

// await polls until the backend answered the request. Each poll waits for messages, thus this never spins.
func (c *wireClient) await(rid proto.RID) error {
	deadline := time.Now().Add(nagotest.SettleTimeout)
	for !c.answered(rid) {
		if c.reloaded() {
			return fmt.Errorf("received a request to reload the page instead of the answer to request %d: %v", rid, eventNames(c.events))
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("no answer to request %d within %v, received %v", rid, nagotest.SettleTimeout, eventNames(c.events))
		}

		if _, err := c.poll(); err != nil {
			return err
		}
	}

	return nil
}

// httpAttempt connects to the window id through the http wire of the session like the frontend: the poll without a
// channel connects to the window, then the client requests the page and polls until it is answered. All its requests
// are completed on return.
func httpAttempt(t *testing.T, base string, sid session.ID, scope proto.ScopeID) wireResult {
	t.Helper()

	c := newWireClient(t, base, sid, scope)
	res := wireResult{close: func() {}}
	if _, err := c.poll(); err != nil {
		res.err = err
		return res
	}

	if c.reloaded() {
		res.events = c.events
		return res
	}

	status, err := c.send(t, pageRequests()...)
	if err == nil && status != http.StatusNoContent {
		err = fmt.Errorf("the send answered %d", status)
	}

	if err == nil {
		err = c.await(ridPage)
	}

	res.events, res.err = c.events, err
	return res
}

type attempt func(t *testing.T, base string, sid session.ID, scope proto.ScopeID) wireResult

var wires = []struct {
	name    string
	dial    dialer
	attempt attempt
}{
	{name: "websocket", dial: nagotest.Dial, attempt: wsAttempt},
	{name: "http", dial: nagotest.DialHTTP, attempt: httpAttempt},
}

// the sessions of the tests: P owns the window, P2 is another session of the same person, Q another person, Anon a
// session without login
var (
	sidP    = session.ID(strings.Repeat("p", 40))
	sidP2   = session.ID(strings.Repeat("r", 40))
	sidQ    = session.ID(strings.Repeat("q", 40))
	sidAnon = session.ID(strings.Repeat("n", 40))
)

func loginOwners(s *linkServer) {
	p := s.person("p@example.com")
	s.login(sidP, p)
	s.login(sidP2, p)
	s.login(sidQ, s.person("q@example.com"))
}

type eolWant int

const (
	eolKept eolWant = iota
	eolExtended
	eolUnasserted // the contract says nothing about it
)

func expectEOL(t *testing.T, row string, before, after time.Time, want eolWant) {
	t.Helper()

	switch {
	case want == eolKept && !after.Equal(before):
		t.Errorf("%s: the end of life of the window moved from %v to %v, want it kept", row, before, after)
	case want == eolExtended && !after.After(before):
		t.Errorf("%s: the end of life of the window stayed at %v, want it extended", row, before)
	}
}

// A connection to the id of an existing window shows that window only to the session of the window. Any other session,
// logged in or not, and a request without cookie get only the request to reload the page, while the window's own
// connection stays usable and its state unchanged. The same connection by the window's session gets the person and the
// state of the window, and an unknown id gets a new window for any session.
func TestWindowConnectsOnlyItsSession(t *testing.T) {
	for _, tr := range wires {
		t.Run(tr.name, func(t *testing.T) {
			s := serveLinks(t, func(cfg *application.Configurator) {
				cfg.RootView(".", windowPage)
			})
			loginOwners(s)

			wp := tr.dial(t, s.base, ".", nagotest.Session(string(sidP)))
			wp.Find(nagotest.Text("user: p@example.com"))
			wp.Click(wp.Find(nagotest.Text("increment")))
			wp.Find(nagotest.Text("count: 1"))
			owned := wp.ScopeID()

			// the rows run in order: the refusals come first, then the window's own connection shows that it is still
			// usable and unchanged, last the window's session connects anew like a tab after a lost connection
			rows := []struct {
				name    string
				before  func()
				session session.ID // empty sends no cookie
				scope   proto.ScopeID
				want    []string // the page; nil expects only the request to reload the page
			}{
				{name: "another session of the same person", session: sidP2, scope: owned},
				{name: "another session, logged in as Q", session: sidQ, scope: owned},
				{name: "another session without login", session: sidAnon, scope: owned},
				{name: "no cookie", scope: owned},
				{name: "unknown id, another session, logged in as Q", session: sidQ, scope: proto.NewScopeID(), want: []string{"user: q@example.com", "count: 0"}},
				{name: "unknown id, another session without login", session: sidAnon, scope: proto.NewScopeID(), want: []string{"user: anonymous", "count: 0"}},
				{name: "unknown id of 32 characters, no cookie", scope: proto.ScopeID(strings.Repeat("u", 32)), want: []string{"user: anonymous", "count: 0"}},
				{name: "the session of the window", before: func() {
					// the window's own connection did not reconnect and still has its person and its state
					wp.Find(nagotest.Text("user: p@example.com"))
					wp.Click(wp.Find(nagotest.Text("increment")))
					wp.Find(nagotest.Text("count: 2"))
					wp.Close()
				}, session: sidP, scope: owned, want: []string{"user: p@example.com", "count: 2"}},
			}

			for _, row := range rows {
				if row.before != nil {
					row.before()
				}

				res := tr.attempt(t, s.base, row.session, row.scope)
				if row.want == nil {
					if err := res.onlyReload(); err != nil {
						t.Errorf("%s: %v", row.name, err)
					}

					continue
				}

				res.expectPage(t, row.name, row.want)
			}
		})
	}

	// proves NAGO-WINDOW-OWNER
}

// A refused connection to a window does not extend its end of life, the same connection by the window's session does.
func TestRefusedWindowConnectKeepsEndOfLife(t *testing.T) {
	for _, tr := range wires {
		t.Run(tr.name, func(t *testing.T) {
			s := serveLinks(t, func(cfg *application.Configurator) {
				cfg.RootView(".", windowPage)
			})
			loginOwners(s)

			// the window's session creates the window; afterward, no request of it is pending, thus neither a message
			// nor a poll nor a keepalive moves the end of life in between
			owned := proto.NewScopeID()
			created := tr.attempt(t, s.base, sidP, owned)
			created.expectPage(t, "the session creating the window", []string{"user: p@example.com"})
			scope, ok := s.app.Scope(owned)
			if !ok {
				t.Fatal("the connection created no window")
			}

			rows := []struct {
				name    string
				before  func()
				session session.ID // empty sends no cookie
				eol     eolWant
			}{
				{name: "another session of the same person", session: sidP2, eol: eolKept},
				{name: "another session, logged in as Q", session: sidQ, eol: eolKept},
				{name: "another session without login", session: sidAnon, eol: eolKept},
				{name: "no cookie", eol: eolKept},
				{name: "the session of the window", before: created.close, session: sidP, eol: eolExtended},
			}

			for _, row := range rows {
				if row.before != nil {
					row.before()
				}

				before := scope.EOL()
				tr.attempt(t, s.base, row.session, owned) // what the client received is the subject of TestWindowConnectsOnlyItsSession
				expectEOL(t, row.name, before, scope.EOL(), row.eol)
			}
		})
	}

	// proves NAGO-WINDOW-OWNER
}

// The token of a channel of the http wire does not grant its window to another session or a request without cookie:
// a send with it answers 403 and does not reach the window, a poll with it gets only the request to reload the page.
// Neither changes the channel or the end of life of the window, while the same requests of the window's session
// work before and after the refusals. A send without a valid token answers 409.
func TestWireTokenNeedsItsSession(t *testing.T) {
	s := serveLinks(t, func(cfg *application.Configurator) {
		cfg.RootView(".", windowPage)
	})
	loginOwners(s)

	// the window's session creates the window and learns the token of its channel
	owner := newWireClient(t, s.base, sidP, proto.NewScopeID())
	if _, err := owner.poll(); err != nil {
		t.Fatalf("the window's session cannot poll: %v", err)
	}

	token := owner.token
	if token == "" {
		t.Fatal("the poll answered without a channel")
	}

	if status, err := owner.send(t, pageRequests()...); err != nil || status != http.StatusNoContent {
		t.Fatalf("the window's session cannot send: status %d, %v", status, err)
	}

	if err := owner.await(ridPage); err != nil {
		t.Fatal(err)
	}

	scope, ok := s.app.Scope(owner.scope)
	if !ok {
		t.Fatal("the poll created no window")
	}

	unknownToken := strings.Repeat("0", len(token))
	if unknownToken == token {
		unknownToken = strings.Repeat("1", len(token))
	}

	// ownSend requests a render as the window's session and waits for its answer, which the owner leaves
	// unacknowledged: the channel always holds a message, which a poll with ack 0 would get at once
	ownSend := func(rid proto.RID) func(row string) {
		return func(row string) {
			status, err := owner.send(t, &proto.RootViewRenderingRequested{RID: rid})
			if err != nil || status != http.StatusNoContent {
				t.Errorf("%s: the send answered %d (%v), want %d", row, status, err, http.StatusNoContent)
				return
			}

			if err := owner.await(rid); err != nil {
				t.Errorf("%s: %v", row, err)
			}
		}
	}

	// foreignSend requests a render with the token as the given session, numbered as the next message of the channel,
	// thus the window would take it, if the session did not count
	foreignSend := func(sid session.ID, token string, rid proto.RID, want int) func(row string) {
		return func(row string) {
			c := owner.as(t, sid)
			c.token = token
			status, err := c.send(t, &proto.RootViewRenderingRequested{RID: rid})
			if err != nil || status != want {
				t.Errorf("%s: the send answered %d (%v), want %d", row, status, err, want)
			}
		}
	}

	// foreignPoll polls the channel with ack 0 as the given session, which would get the unacknowledged messages of
	// the window at once, if the session did not count
	foreignPoll := func(sid session.ID) func(row string) {
		return func(row string) {
			c := owner.as(t, sid)
			channel, err := c.poll()
			if err != nil {
				t.Errorf("%s: %v", row, err)
				return
			}

			if channel == token {
				t.Errorf("%s: the answer names the channel of the window", row)
			}

			if err := (wireResult{events: c.events}).onlyReload(); err != nil {
				t.Errorf("%s: %v", row, err)
			}
		}
	}

	refused := []proto.RID{2002, 2003, 2004, 2005}

	// the rows run in order, the window's session sends before and after the refusals
	rows := []struct {
		name string
		act  func(row string)
		eol  eolWant
	}{
		{name: "send with the token, the window's session", act: ownSend(2001), eol: eolExtended},
		{name: "send with the token, another session of the same person", act: foreignSend(sidP2, token, refused[0], http.StatusForbidden), eol: eolKept},
		{name: "send with the token, another session, logged in as Q", act: foreignSend(sidQ, token, refused[1], http.StatusForbidden), eol: eolKept},
		{name: "send with the token, no cookie", act: foreignSend("", token, refused[2], http.StatusForbidden), eol: eolKept},
		{name: "send with an unknown token, the window's session", act: foreignSend(sidP, unknownToken, refused[3], http.StatusConflict), eol: eolUnasserted},
		{name: "poll with the token, another session of the same person", act: foreignPoll(sidP2), eol: eolKept},
		{name: "poll with the token, another session, logged in as Q", act: foreignPoll(sidQ), eol: eolKept},
		{name: "poll with the token, no cookie", act: foreignPoll(""), eol: eolKept},
		// the same number as the refused sends: the window took none of them, and its channel is still usable
		{name: "send and poll with the token, the window's session after the refusals", act: ownSend(2006), eol: eolExtended},
	}

	for _, row := range rows {
		before := scope.EOL()
		row.act(row.name)
		expectEOL(t, row.name, before, scope.EOL(), row.eol)
	}

	// the window processes in order, thus the answer to the last send follows the answer to any refused one
	for _, rid := range refused {
		if owner.answered(rid) {
			t.Errorf("the window answered the refused request %d: %v", rid, eventNames(owner.events))
		}
	}

	// proves NAGO-WINDOW-OWNER
}

// uploadAs posts the content like the frontend into the import of the window as the given session (none for an empty
// id) and returns the answer.
func uploadAs(t *testing.T, base string, sid session.ID, window proto.ScopeID, importID, content string) (int, []byte) {
	t.Helper()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("data.txt", "data.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}

	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	header := http.Header{
		"Content-Type": {mw.FormDataContentType()},
		"X-Scope":      {string(window)},
		"X-Receiver":   {url.PathEscape(importID)}, // like encodeURIComponent of the frontend
	}

	status, _, resp, err := newBrowser(t, base, sid).post("/api/ora/v1/upload", nil, header, body.Bytes())
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	return status, resp
}

// An upload into a pending import of a window is accepted only from the session of the window. Any other request gets
// the answer to an unknown upload, the import is not completed and the end of life of the window is kept, while the
// same upload of the window's session completes the import.
func TestUploadOnlyByItsSession(t *testing.T) {
	const importID = "import-4711"

	var mutex sync.Mutex
	var received []string
	completed := make(chan struct{}, 16) // large enough, that a wrong completion never blocks

	s := serveLinks(t, func(cfg *application.Configurator) {
		cfg.RootView(".", func(wnd core.Window) core.View {
			pings := core.AutoState[int](wnd)
			return ui.VStack(
				ui.Text("user: "+who(wnd)),
				ui.PrimaryButton(func() {
					wnd.ImportFiles(core.ImportFilesOptions{ID: importID, OnCompletion: func(files []core.File) {
						for _, file := range files {
							var buf bytes.Buffer
							_, _ = file.Transfer(&buf)
							mutex.Lock()
							received = append(received, buf.String())
							mutex.Unlock()
						}

						completed <- struct{}{}
					}})
				}).Title("import"),
				ui.PrimaryButton(func() { pings.Set(pings.Get() + 1) }).Title("ping"),
				ui.Text(fmt.Sprintf("pings: %d", pings.Get())),
			)
		})
	})
	loginOwners(s)

	// the window's connection stays open and idle, a keepalive is far away
	wp := nagotest.Dial(t, s.base, ".", nagotest.Session(string(sidP)))
	wp.Find(nagotest.Text("user: p@example.com"))
	wp.Click(wp.Find(nagotest.Text("import")))
	owned := wp.ScopeID()
	imports := wp.Imports()
	if len(imports) != 1 || string(imports[0].ID) != importID || proto.ScopeID(imports[0].ScopeID) != owned {
		t.Fatalf("want one import %q of the window %q, got %+v", importID, owned, imports)
	}

	scope, ok := s.app.Scope(owned)
	if !ok {
		t.Fatal("the window has no scope")
	}

	// the answer to an unknown upload, which every refusal must repeat
	unknownStatus, unknownBody := uploadAs(t, s.base, sidP, owned, "unknown-import", "unknown import")
	if unknownStatus/100 == 2 {
		t.Fatalf("an unknown upload must be refused, got %d with %q", unknownStatus, unknownBody)
	}

	refusal := func(sid session.ID, window proto.ScopeID) func(row string) {
		return func(row string) {
			status, body := uploadAs(t, s.base, sid, window, importID, "from "+row)
			if status != unknownStatus || !bytes.Equal(body, unknownBody) {
				t.Errorf("%s: want the answer to an unknown upload %d with %q, got %d with %q", row, unknownStatus, unknownBody, status, body)
			}
		}
	}

	// the rows run in order: the refusals come first, then the window's session completes the import, last a message
	// of the window shows that the end of life does move
	rows := []struct {
		name string
		act  func(row string)
		eol  eolWant
	}{
		{name: "another session of the same person", act: refusal(sidP2, owned), eol: eolKept},
		{name: "another session, logged in as Q", act: refusal(sidQ, owned), eol: eolKept},
		{name: "another session without login", act: refusal(sidAnon, owned), eol: eolKept},
		{name: "no cookie", act: refusal("", owned), eol: eolKept},
		{name: "unknown window, the window's session", act: refusal(sidP, proto.NewScopeID()), eol: eolKept},
		{name: "the window's session", act: func(row string) {
			if status, body := uploadAs(t, s.base, sidP, owned, importID, "from "+row); status != http.StatusOK {
				t.Errorf("%s: want 200, got %d with %q", row, status, body)
				return
			}

			select {
			case <-completed:
			case <-time.After(nagotest.SettleTimeout):
				t.Errorf("%s: the import was not completed within %v", row, nagotest.SettleTimeout)
			}
		}, eol: eolUnasserted},
		{name: "message of the window", act: func(string) {
			wp.Click(wp.Find(nagotest.Text("ping")))
			wp.Find(nagotest.Text("pings: 1"))
		}, eol: eolExtended},
	}

	for _, row := range rows {
		before := scope.EOL()
		row.act(row.name)
		expectEOL(t, row.name, before, scope.EOL(), row.eol)
	}

	mutex.Lock()
	defer mutex.Unlock()
	if want := []string{"from the window's session"}; !slices.Equal(received, want) {
		t.Errorf("the import received %q, want %q", received, want)
	}

	// proves NAGO-UPLOAD-OWNER
}
