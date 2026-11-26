// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/core/http/poll"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

func configureHTTP(cfg *application.Configurator) {
	cfg.SetApplicationID("de.worldiety.nagotest.http")

	cfg.RootView("counter", func(wnd core.Window) core.View {
		count := core.AutoState[int](wnd)
		name := core.AutoState[string](wnd)
		greeting := core.AutoState[string](wnd)

		return ui.VStack(
			ui.Text(fmt.Sprintf("count: %d", count.Get())),
			ui.PrimaryButton(func() { count.Set(count.Get() + 1) }).Title("increment"),
			ui.TextField("Name", name.Get()).InputValue(name).KeydownEnter(func() {
				greeting.Set("hello " + name.Get())
			}),
			ui.Text(greeting.Get()),
			ui.PrimaryButton(func() {
				wnd.ExportFiles(core.ExportFileBytes("a.txt", []byte("hello")))
			}).Title("download"),
			ui.PrimaryButton(func() {
				wnd.Navigation().ForwardTo("details", core.Values{"id": "42"})
			}).Title("details"),
		)
	})

	cfg.RootView("details", func(wnd core.Window) core.View {
		return ui.Text("details of " + wnd.Values()["id"])
	})

	// the backend renders by itself, while the window waits within a poll
	cfg.RootView("late", func(wnd core.Window) core.View {
		late := core.AutoState[string](wnd)
		core.OnAppear(wnd, "", func(ctx context.Context) {
			time.Sleep(200 * time.Millisecond)
			late.Set("rendered later")
		})

		return ui.Text(late.Get())
	})
}

// A window works over the http fallback of the wire like over the websocket, including a reconnect, which keeps the
// states.
func TestDialHTTP(t *testing.T) {
	base := nagotest.Serve(t, configureHTTP)
	w := nagotest.DialHTTP(t, base, "counter")

	w.Find(nagotest.Text("count: 0"))
	w.Click(w.Find(nagotest.Text("increment")))
	w.Click(w.Find(nagotest.Text("increment")))
	w.Find(nagotest.Text("count: 2"))

	w.Type(w.Find(nagotest.Label("Name")), "Torben")
	w.PressEnter(w.Find(nagotest.Type[*proto.TextField]()))
	w.Find(nagotest.Text("hello Torben"))

	w.Click(w.Find(nagotest.Text("download")))
	if r := w.Resources(); len(r) != 1 {
		t.Fatalf("unexpected resources %v", r)
	}

	w.Reconnect()
	w.Find(nagotest.Text("count: 2"))
	w.Click(w.Find(nagotest.Text("increment")))
	w.Find(nagotest.Text("count: 3"))

	w.Click(w.Find(nagotest.Text("details")))
	w.Find(nagotest.Text("details of 42"))

	nagotest.DialHTTP(t, base, "late").WaitFor(nagotest.Text("rendered later"), 5*time.Second)
}

// TestHTTPWireRequests checks the requests of the http wire, which a browser does not send.
func TestHTTPWireRequests(t *testing.T) {
	base := nagotest.Serve(t, configureHTTP)
	sid := strings.Repeat("a", 64)

	// the client stands for the browser of the window: from the first poll on, it keeps the cookie of the session, which
	// that poll issues, and sends it with the token like the frontend
	client := &http.Client{}

	do := func(method, path string, query url.Values, header http.Header, body []byte) *http.Response {
		t.Helper()
		u, _ := url.Parse(base)
		u.Path = path
		u.RawQuery = query.Encode()
		req, err := http.NewRequest(method, u.String(), bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}

		for k, v := range header {
			req.Header[k] = v
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() { _ = resp.Body.Close() })
		return resp
	}

	expect := func(resp *http.Response, status int) {
		t.Helper()
		if resp.StatusCode != status {
			msg, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected status %d, got %d: %s", status, resp.StatusCode, msg)
		}
	}

	expect(do(http.MethodGet, "/wire/poll", url.Values{"_sid": {sid}}, nil, nil), http.StatusMethodNotAllowed)
	expect(do(http.MethodPost, "/wire/poll", url.Values{"_sid": {"short"}}, nil, nil), http.StatusBadRequest)
	expect(do(http.MethodPost, "/wire/poll", url.Values{"_sid": {sid}}, http.Header{"Origin": {"https://evil.example"}}, nil), http.StatusForbidden)
	expect(do(http.MethodPost, "/wire/send", url.Values{"_sid": {sid}, "_c": {"unknown"}}, nil, nil), http.StatusConflict)

	// the requests above are no requests of the window; its browser starts without a cookie, thus the first poll must
	// issue one, which the jar then sends with every request carrying the token
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client.Jar = jar

	// a new channel answers at once with its token and the cookie of a new session
	start := time.Now()
	resp := do(http.MethodPost, "/wire/poll", url.Values{"_sid": {sid}}, nil, nil)
	expect(resp, http.StatusOK)
	token := resp.Header.Get("X-Nago-Channel")
	if token == "" || len(resp.Cookies()) == 0 || time.Since(start) > 5*time.Second {
		t.Fatalf("expected a token and a session cookie at once, got %q %v after %v", token, resp.Cookies(), time.Since(start))
	}

	if cc := resp.Header.Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("a poll must not be cached: %q", cc)
	}

	// a poll of a known channel waits for messages
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	u, _ := url.Parse(base)
	u.Path = "/wire/poll"
	u.RawQuery = url.Values{"_sid": {sid}, "_c": {token}}.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if resp, err := client.Do(req); err == nil {
		_ = resp.Body.Close()
		t.Fatalf("expected the poll to wait, got status %d", resp.StatusCode)
	}

	// a message, which is no event, breaks the channel like a websocket, thus the client needs a new one
	var body bytes.Buffer
	_ = poll.WriteFrames(&body, 1, [][]byte{{0xff, 0xff, 0xff}})
	expect(do(http.MethodPost, "/wire/send", url.Values{"_sid": {sid}, "_c": {token}}, nil, []byte("\xff")), http.StatusBadRequest)
	expect(do(http.MethodPost, "/wire/send", url.Values{"_sid": {sid}, "_c": {token}}, nil, body.Bytes()), http.StatusConflict)
	expect(do(http.MethodPost, "/wire/send", url.Values{"_sid": {sid}, "_c": {token}}, nil, body.Bytes()), http.StatusConflict)

	resp = do(http.MethodPost, "/wire/poll", url.Values{"_sid": {sid}, "_c": {token}}, nil, nil)
	expect(resp, http.StatusOK)
	if next := resp.Header.Get("X-Nago-Channel"); next == "" || next == token {
		t.Fatalf("expected a new channel, got %q", next)
	}
}
