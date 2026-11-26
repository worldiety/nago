// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/web/vuejs"
)

// envChrome names the Chrome binary for TestEmbeddedPageInChrome, e.g.
//
//	NAGO_CHROME="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" go test -run TestEmbeddedPageInChrome ./application
const envChrome = "NAGO_CHROME"

// browserWait bounds each wait for the browser, which starts, loads the frontend and connects.
const browserWait = 30 * time.Second

// served is a request, which the server of the application answered.
type served struct {
	path    string
	query   string
	status  int
	cookies []string // the names of the cookies of the request
}

// recorder records the requests of the application and keeps the connections of its websockets, to interrupt them.
type recorder struct {
	mutex   sync.Mutex
	served  []served
	conns   []net.Conn
	changed chan struct{}
}

func newRecorder() *recorder {
	return &recorder{changed: make(chan struct{})}
}

func (rec *recorder) add(r *http.Request, status int, conn net.Conn) {
	var names []string
	for _, c := range r.Cookies() {
		names = append(names, c.Name)
	}

	rec.mutex.Lock()
	defer rec.mutex.Unlock()

	rec.served = append(rec.served, served{path: r.URL.Path, query: r.URL.RawQuery, status: status, cookies: names})
	if conn != nil {
		rec.conns = append(rec.conns, conn)
	}

	close(rec.changed)
	rec.changed = make(chan struct{})
}

// count returns the number of requests of the path.
func (rec *recorder) count(path string) int {
	rec.mutex.Lock()
	defer rec.mutex.Unlock()

	n := 0
	for _, s := range rec.served {
		if s.path == path {
			n++
		}
	}

	return n
}

// awaitNth waits for the n-th request of the path, counted from 1.
func (rec *recorder) awaitNth(t *testing.T, path string, n int) served {
	t.Helper()

	deadline := time.After(browserWait)
	for {
		rec.mutex.Lock()
		var found []served
		for _, s := range rec.served {
			if s.path == path {
				found = append(found, s)
			}
		}
		changed := rec.changed
		rec.mutex.Unlock()

		if len(found) >= n {
			return found[n-1]
		}

		select {
		case <-changed:
		case <-deadline:
			t.Fatalf("the application received no request %d of %s within %v", n, path, browserWait)
		}
	}
}

// websockets returns the number of the connections of websockets so far.
func (rec *recorder) websockets() int {
	rec.mutex.Lock()
	defer rec.mutex.Unlock()

	return len(rec.conns)
}

// interrupt closes the connections of the websockets from the given number on, like a lost network.
func (rec *recorder) interrupt(from int) {
	rec.mutex.Lock()
	defer rec.mutex.Unlock()

	for _, conn := range rec.conns[from:] {
		_ = conn.Close()
	}
}

func (rec *recorder) wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &recordingWriter{ResponseWriter: w, rec: rec, r: r}
		h.ServeHTTP(rw, r)
		if !rw.hijacked {
			rec.add(r, max(rw.status, http.StatusOK), nil)
		}
	})
}

type recordingWriter struct {
	http.ResponseWriter
	rec      *recorder
	r        *http.Request
	status   int
	hijacked bool
}

func (w *recordingWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}

	w.ResponseWriter.WriteHeader(status)
}

func (w *recordingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, brw, err := http.NewResponseController(w.ResponseWriter).Hijack()
	if err == nil {
		w.hijacked = true
		w.rec.add(w.r, http.StatusSwitchingProtocols, conn)
	}

	return conn, brw, err
}

func (w *recordingWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// cdpMessage is a message of the Chrome DevTools Protocol.
type cdpMessage struct {
	ID        int                       `json:"id,omitempty"`
	Method    string                    `json:"method,omitempty"`
	Params    json.RawMessage           `json:"params,omitempty"`
	Result    json.RawMessage           `json:"result,omitempty"`
	Error     *struct{ Message string } `json:"error,omitempty"`
	SessionID string                    `json:"sessionId,omitempty"`
}

// cdp is a client of the Chrome DevTools Protocol, just enough to drive a page and its frames.
type cdp struct {
	conn    *websocket.Conn
	wmutex  sync.Mutex
	mutex   sync.Mutex
	next    int
	pending map[int]chan cdpMessage
	events  []cdpMessage
	changed chan struct{}
}

// startChrome starts a headless Chrome with a profile of its own and connects to it. Chrome ends with the test.
func startChrome(t *testing.T, binary string) *cdp {
	t.Helper()

	profile := t.TempDir()
	cmd := exec.Command(binary,
		"--headless=new",
		"--remote-debugging-port=0",
		"--user-data-dir="+profile,
		"--no-first-run",
		"--no-default-browser-check",
		"--site-per-process", // a frame of another site is a target of its own
		// the frontend does not start for a crawler, whose list names HeadlessChrome
		"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36",
		"about:blank",
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("cannot start %s: %v", binary, err)
	}

	exited := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(exited)
	}()

	kill := func() {
		_ = cmd.Process.Kill()
		<-exited
	}

	// Chrome tells its port in the profile, once it listens
	var endpoint string
	deadline := time.Now().Add(browserWait)
	for endpoint == "" {
		select {
		case <-exited:
			t.Fatal("Chrome exited before it listened")
		default:
		}

		if buf, err := os.ReadFile(filepath.Join(profile, "DevToolsActivePort")); err == nil {
			if lines := strings.Split(strings.TrimSpace(string(buf)), "\n"); len(lines) == 2 {
				endpoint = "ws://127.0.0.1:" + lines[0] + lines[1]
				break
			}
		}

		if time.Now().After(deadline) {
			kill()
			t.Fatalf("Chrome did not listen within %v", browserWait)
		}

		time.Sleep(50 * time.Millisecond)
	}

	conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		kill()
		t.Fatalf("cannot connect to Chrome: %v", err)
	}

	c := &cdp{conn: conn, pending: map[int]chan cdpMessage{}, changed: make(chan struct{})}
	go c.read()
	t.Cleanup(func() {
		// unlike a kill, closing the browser also ends its helper processes at once
		_, _ = c.call("", "Browser.close", map[string]any{})
		select {
		case <-exited:
		case <-time.After(browserWait):
			kill()
		}

		_ = conn.Close()
	})

	return c
}

func (c *cdp) read() {
	for {
		var msg cdpMessage
		if err := c.conn.ReadJSON(&msg); err != nil {
			c.mutex.Lock()
			for id, ch := range c.pending {
				close(ch)
				delete(c.pending, id)
			}
			c.mutex.Unlock()
			return
		}

		c.mutex.Lock()
		if msg.ID != 0 {
			if ch, ok := c.pending[msg.ID]; ok {
				ch <- msg
				delete(c.pending, msg.ID)
			}
		} else {
			c.events = append(c.events, msg)
			close(c.changed)
			c.changed = make(chan struct{})
		}
		c.mutex.Unlock()
	}
}

// call invokes the method within the session, or the browser for an empty one, and returns its result.
func (c *cdp) call(session, method string, params any) (json.RawMessage, error) {
	buf, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	ch := make(chan cdpMessage, 1)
	c.mutex.Lock()
	c.next++
	id := c.next
	c.pending[id] = ch
	c.mutex.Unlock()

	c.wmutex.Lock()
	err = c.conn.WriteJSON(cdpMessage{ID: id, Method: method, Params: buf, SessionID: session})
	c.wmutex.Unlock()
	if err != nil {
		return nil, err
	}

	select {
	case msg, ok := <-ch:
		switch {
		case !ok:
			return nil, errors.New("the connection to Chrome is gone")
		case msg.Error != nil:
			return nil, fmt.Errorf("%s: %s", method, msg.Error.Message)
		}

		return msg.Result, nil
	case <-time.After(browserWait):
		return nil, fmt.Errorf("%s: no answer within %v", method, browserWait)
	}
}

func (c *cdp) must(t *testing.T, session, method string, params any) json.RawMessage {
	t.Helper()

	res, err := c.call(session, method, params)
	if err != nil {
		t.Fatal(err)
	}

	return res
}

// awaitEvent waits for the event of the method, whose parameters match, and returns its parameters.
func (c *cdp) awaitEvent(t *testing.T, method string, match func(params json.RawMessage) bool) json.RawMessage {
	t.Helper()

	deadline := time.After(browserWait)
	for seen := 0; ; {
		c.mutex.Lock()
		events := c.events[seen:]
		seen = len(c.events)
		changed := c.changed
		c.mutex.Unlock()

		for _, evt := range events {
			if evt.Method == method && match(evt.Params) {
				return evt.Params
			}
		}

		select {
		case <-changed:
		case <-deadline:
			t.Fatalf("Chrome sent no matching %s within %v", method, browserWait)
		}
	}
}

// awaitFrame waits for the frame of another site within the page and returns its session. Chrome attaches it before
// its navigation commits, thus it is found by its parent, not by its address.
func (c *cdp) awaitFrame(t *testing.T, page string) string {
	t.Helper()

	var attached struct {
		SessionID  string `json:"sessionId"`
		TargetInfo struct {
			Type     string `json:"type"`
			ParentID string `json:"parentId"`
		} `json:"targetInfo"`
	}

	c.awaitEvent(t, "Target.attachedToTarget", func(params json.RawMessage) bool {
		return json.Unmarshal(params, &attached) == nil && attached.TargetInfo.Type == "iframe" && attached.TargetInfo.ParentID == page
	})

	return attached.SessionID
}

// open opens a tab of the browser context, empty for the default one, at the address and returns its target and its
// session.
func (c *cdp) open(t *testing.T, context, address string) (target, session string) {
	t.Helper()

	params := map[string]any{"url": "about:blank"}
	if context != "" {
		params["browserContextId"] = context
	}

	var created struct {
		TargetID string `json:"targetId"`
	}

	if err := json.Unmarshal(c.must(t, "", "Target.createTarget", params), &created); err != nil {
		t.Fatal(err)
	}

	var attached struct {
		SessionID string `json:"sessionId"`
	}

	if err := json.Unmarshal(c.must(t, "", "Target.attachToTarget", map[string]any{"targetId": created.TargetID, "flatten": true}), &attached); err != nil {
		t.Fatal(err)
	}

	c.must(t, attached.SessionID, "Target.setAutoAttach", map[string]any{"autoAttach": true, "waitForDebuggerOnStart": false, "flatten": true})
	c.must(t, attached.SessionID, "Page.navigate", map[string]any{"url": address})
	return created.TargetID, attached.SessionID
}

// eval evaluates the expression in the frame like a click of the user and returns its value as JSON.
func (c *cdp) eval(session, expr string) (json.RawMessage, error) {
	res, err := c.call(session, "Runtime.evaluate", map[string]any{"expression": expr, "returnByValue": true, "awaitPromise": true, "userGesture": true})
	if err != nil {
		return nil, err
	}

	var out struct {
		Result struct {
			Value json.RawMessage `json:"value"`
		} `json:"result"`
		ExceptionDetails *struct {
			Text string `json:"text"`
		} `json:"exceptionDetails"`
	}

	if err := json.Unmarshal(res, &out); err != nil {
		return nil, err
	}

	if out.ExceptionDetails != nil {
		return nil, fmt.Errorf("%s: %s", expr, out.ExceptionDetails.Text)
	}

	return out.Result.Value, nil
}

// text returns the visible text of the frame.
func (c *cdp) text(session string) (string, error) {
	v, err := c.eval(session, `document.body ? document.body.innerText : ""`)
	if err != nil {
		return "", err
	}

	var s string
	err = json.Unmarshal(v, &s)
	return s, err
}

// awaitText waits until the frame shows the text. A frame, which loads again in between, is asked again.
func (c *cdp) awaitText(t *testing.T, session, want string) {
	t.Helper()

	deadline := time.Now().Add(browserWait)
	for {
		got, err := c.text(session)
		if err == nil && strings.Contains(got, want) {
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("the frame did not show %q within %v, it shows %q (%v)", want, browserWait, got, err)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// click clicks the button of the frame with the title.
func (c *cdp) click(t *testing.T, session, title string) {
	t.Helper()

	expr := fmt.Sprintf(`(() => {
		const b = [...document.querySelectorAll("button")].find(b => b.innerText.trim() === %q);
		if (!b) { throw new Error("no button " + %q); }
		b.click();
		return true;
	})()`, title, title)
	if _, err := c.eval(session, expr); err != nil {
		t.Fatal(err)
	}
}

// A page in a frame of another site, which is embeddable, downloads and connects again after an interruption of its
// websocket without loading again in Chrome. Its embedded cookie is SameSite=None, Secure and Partitioned for the
// embedding site. The same link direct is not found. In a browser context without the embedded cookie, the same page
// under a path, which is not embeddable, gets no cookie in the frame: its download is not found, and the
// interruption loads it again, which shows that the embedded cookie makes the difference. The application runs on
// localhost, the embedding site on 127.0.0.1.
func TestEmbeddedPageInChrome(t *testing.T) {
	binary := os.Getenv(envChrome)
	if binary == "" {
		t.Skipf("set %s to the binary of Chrome", envChrome)
	}

	rec := newRecorder()
	cfg := application.NewConfigurator()
	cfg.SetDataDir(t.TempDir())
	cfg.SetApplicationID("de.worldiety.embeddedchrometest")
	std.Must(cfg.SessionManagement())
	cfg.Serve(vuejs.Dist())
	cfg.EmbeddablePages("e/*")
	cfg.RootView("e/*", embeddedPage)
	cfg.RootView("x/*", embeddedPage)
	handler, app, err := cfg.NewHTTPHandler()
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(rec.wrap(handler))
	t.Cleanup(func() {
		srv.CloseClientConnections()
		srv.Close()
		app.Destroy()
	})

	appURL := strings.Replace(srv.URL, "127.0.0.1", "localhost", 1)
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<!doctype html><title>another site</title><iframe src="%s%s" style="width:900px;height:700px"></iframe>`, appURL, r.URL.Path)
	}))
	t.Cleanup(site.Close)

	chrome := startChrome(t, binary)
	downloads := t.TempDir()
	chrome.must(t, "", "Browser.setDownloadBehavior", map[string]any{"behavior": "allowAndName", "downloadPath": downloads, "eventsEnabled": true})

	type outcome struct {
		download int    // the status of the download
		cookies  string // the cookies of the download
		kept     bool   // the frame kept its document and its window after the interruption
	}

	// embed opens the page in the frame of another site in the browser context, counts, downloads and interrupts the
	// websocket of the frame
	embed := func(context, path string) (outcome, served) {
		conns := rec.websockets()
		page, _ := chrome.open(t, context, site.URL+path)
		frame := chrome.awaitFrame(t, page)
		chrome.awaitText(t, frame, "count: 0")
		chrome.click(t, frame, "increment")
		chrome.awaitText(t, frame, "count: 1")
		if _, err := chrome.eval(frame, `window.nagoMarker = "kept"`); err != nil {
			t.Fatal(err)
		}

		n := rec.count("/api/ora/v1/download")
		chrome.click(t, frame, "export")
		download := rec.awaitNth(t, "/api/ora/v1/download", n+1)
		res := outcome{download: download.status, cookies: strings.Join(download.cookies, " ")}

		wires := rec.count("/wire")
		rec.interrupt(conns)
		rec.awaitNth(t, "/wire", wires+1)

		// A kept window counts on from 1, a window, which loaded again, starts at 0 without the marker. A click
		// before the websocket is open again is lost, thus it is repeated.
		var clicked time.Time
		deadline := time.Now().Add(browserWait)
		for {
			marker, err := chrome.eval(frame, `window.nagoMarker || ""`)
			got, _ := chrome.text(frame)
			kept := err == nil && string(marker) == `"kept"`
			switch {
			case !kept && err == nil && strings.Contains(got, "count: 0"):
				return res, download
			case kept && !strings.Contains(got, "count: 0") && !strings.Contains(got, "count: 1") && strings.Contains(got, "count: "):
				res.kept = true
				return res, download
			case kept && time.Since(clicked) > 500*time.Millisecond:
				_, _ = chrome.eval(frame, `[...document.querySelectorAll("button")].find(b => b.innerText.trim() === "increment")?.click()`)
				clicked = time.Now()
			case time.Now().After(deadline):
				t.Fatalf("%s: after the interruption, the frame neither kept its window nor loaded again: marker %s, text %q, %v", path, marker, got, err)
			}

			time.Sleep(50 * time.Millisecond)
		}
	}

	embedded, download := embed("", "/e/forms/contact")
	if want := (outcome{download: http.StatusOK, cookies: "wdy-ora-embedded", kept: true}); embedded != want {
		t.Errorf("embeddable page: %+v, want %+v", embedded, want)
	}

	// the download arrived in Chrome
	var progress struct {
		GUID  string `json:"guid"`
		State string `json:"state"`
	}
	chrome.awaitEvent(t, "Browser.downloadProgress", func(params json.RawMessage) bool {
		return json.Unmarshal(params, &progress) == nil && progress.State != "inProgress"
	})

	if got, err := os.ReadFile(filepath.Join(downloads, progress.GUID)); progress.State != "completed" || err != nil || string(got) != string(report("anonymous")) {
		t.Errorf("the download is %s with %q (%v), want completed with %q", progress.State, got, err, report("anonymous"))
	}

	// the premise: in a browser context of its own, the frame of the same site has no embedded cookie of an
	// embeddable page yet, thus the same page under a path, which is not embeddable, gets none
	var context struct {
		BrowserContextID string `json:"browserContextId"`
	}
	if err := json.Unmarshal(chrome.must(t, "", "Target.createBrowserContext", map[string]any{}), &context); err != nil {
		t.Fatal(err)
	}

	notEmbeddable, _ := embed(context.BrowserContextID, "/x/forms/contact")
	if want := (outcome{download: http.StatusNotFound, cookies: "", kept: false}); notEmbeddable != want {
		t.Errorf("page, which is not embeddable: %+v, want %+v (the premise, that a frame of another site needs the embedded cookie)", notEmbeddable, want)
	}

	// the cookie, which Chrome keeps for the embedding site in the default context
	var jar struct {
		Cookies []struct {
			Name         string          `json:"name"`
			Secure       bool            `json:"secure"`
			HTTPOnly     bool            `json:"httpOnly"`
			SameSite     string          `json:"sameSite"`
			PartitionKey json.RawMessage `json:"partitionKey"` // an object with topLevelSite or, before Chrome 125, the site
		} `json:"cookies"`
	}

	if err := json.Unmarshal(chrome.must(t, "", "Storage.getCookies", map[string]any{}), &jar); err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, c := range jar.Cookies {
		names = append(names, c.Name)
		if c.Name != "wdy-ora-embedded" {
			continue
		}

		if !c.Secure || !c.HTTPOnly || c.SameSite != "None" || !strings.Contains(string(c.PartitionKey), `"http://127.0.0.1"`) {
			t.Errorf("want the embedded cookie Secure, HttpOnly, SameSite=None and partitioned for http://127.0.0.1, got %+v with the partition %s", c, c.PartitionKey)
		}
	}

	if !slices.Equal(names, []string{"wdy-ora-embedded"}) {
		t.Errorf("want only the embedded cookie in Chrome, got %v", names)
	}

	// the same link direct, without the cookie of the embedding site
	n := rec.count("/api/ora/v1/download")
	chrome.open(t, "", appURL+"/api/ora/v1/download?"+download.query)
	if direct := rec.awaitNth(t, "/api/ora/v1/download", n+1); direct.status != http.StatusNotFound {
		t.Errorf("the link of the frame direct: status %d, want %d", direct.status, http.StatusNotFound)
	}

	// proves NAGO-EMBED-COOKIE, NAGO-EMBED-SESSION
}
