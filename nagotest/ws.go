// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// RenderTimeout is the duration a websocket [Window] waits for the render of an action, before it requests a
// render by itself. This is only required for the rare actions which change no state.
var RenderTimeout = 2 * time.Second

// KeepAlive is the interval of pings which a websocket [Window] sends, like the frontend.
var KeepAlive = 30 * time.Second

// sessionCookie is the name of the http-only cookie which carries the session id.
const sessionCookie = "wdy-ora-access"

// Dial connects a new window through a websocket to a running nago server at baseURL, e.g.
// http://localhost:3000, and allocates the root view of the given path. It behaves like a browser, thus
// the window is anonymous until it logs in through the user interface. Use [Session] to share the login
// state across windows. The window is closed automatically by t.Cleanup.
//
// Like a browser, an action refers to the tree received last. If the backend renders proactively in the meantime,
// e.g. because a background task of [core.OnAppear] completed, the action is stale and discarded by the backend.
// Use [Window.WaitFor] to await such results before acting.
func Dial(t TB, baseURL string, path core.NavigationPath, opts ...OpenOption) *Window {
	t.Helper()

	base, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("nagotest: invalid url %q: %v", baseURL, err)
	}

	w := newWindow(t, &wsTransport{base: base}, path, opts)
	w.open()
	return w
}

// wsTransport connects a window through a websocket, just like the frontend.
type wsTransport struct {
	base   *url.URL
	mutex  sync.Mutex // guards writes to conn
	conn   *websocket.Conn
	done   chan struct{}
	client http.Client
}

func (c *wsTransport) connect(w *Window) error {
	u := *c.base
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	default:
		u.Scheme = "ws"
	}

	// like the frontend, the wire is always located at the root
	u.Path = "/wire"
	u.RawQuery = url.Values{"_sid": {string(w.scopeID)}}.Encode()

	header := http.Header{}
	header.Set("Cookie", (&http.Cookie{Name: sessionCookie, Value: w.opts.sessionID}).String())

	dialer := websocket.Dialer{EnableCompression: true, HandshakeTimeout: SettleTimeout}
	conn, resp, err := dialer.Dial(u.String(), header)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("%s: %w (status %d)", u.String(), err, resp.StatusCode)
		}

		return fmt.Errorf("%s: %w", u.String(), err)
	}

	done := make(chan struct{})
	c.mutex.Lock()
	c.conn = conn
	c.done = done
	c.mutex.Unlock()

	go c.read(w, conn, done)
	go c.keepAlive(done)
	return nil
}

func (c *wsTransport) read(w *Window, conn *websocket.Conn, done chan struct{}) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			select {
			case <-done:
				// closed intentionally
			default:
				w.fail(fmt.Errorf("connection lost: %w", err))
			}

			return
		}

		obj, err := proto.Unmarshal(proto.NewBinaryReader(bytes.NewBuffer(msg)))
		if err != nil {
			w.fail(fmt.Errorf("cannot decode message: %w", err))
			return
		}

		evt, ok := obj.(proto.NagoEvent)
		if !ok {
			w.fail(unexpected(obj))
			return
		}

		w.receive(evt)
	}
}

func (c *wsTransport) keepAlive(done chan struct{}) {
	ticker := time.NewTicker(KeepAlive)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			_ = c.write(&proto.Ping{})
		}
	}
}

func (c *wsTransport) write(evt proto.NagoEvent) error {
	var buf bytes.Buffer
	if err := proto.Marshal(proto.NewBinaryWriter(&buf), evt); err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn == nil {
		return errors.New("not connected")
	}

	return c.conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
}

func (c *wsTransport) send(w *Window, evt proto.NagoEvent) error {
	return c.write(evt)
}

// settle waits for the answer of the last action. If the action caused only side effects like a navigation,
// or nothing within the RenderTimeout, a render is requested as a barrier. Due to the in-order processing
// of the event loop, its answer proves that the action has been processed entirely.
func (c *wsTransport) settle(w *Window) error {
	w.mutex.Lock()
	rid, mark := w.awaitRID, w.awaitMark
	w.mutex.Unlock()

	if rid == 0 {
		return nil
	}

	var sideEffect bool
	answered := func() bool {
		sideEffect = w.sideEffects > mark
		return w.answered >= rid || sideEffect || w.connErr != nil
	}

	w.waitUntil(RenderTimeout, answered)
	if err := w.connError(); err != nil {
		return err
	}

	w.mutex.Lock()
	done := w.answered >= rid
	w.mutex.Unlock()

	if !done {
		barrier := w.nextRID()
		if err := c.write(&proto.RootViewRenderingRequested{RID: barrier}); err != nil {
			return err
		}

		if !w.waitUntil(SettleTimeout, func() bool { return w.answered >= barrier || w.connErr != nil }) {
			return fmt.Errorf("no answer within %v", SettleTimeout)
		}
	}

	w.mutex.Lock()
	w.awaitRID = 0
	w.mutex.Unlock()

	return w.connError()
}

// upload posts the files to the upload endpoint, just like the frontend.
func (c *wsTransport) upload(w *Window, id string, files []core.File) error {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for i, file := range files {
		part, err := mw.CreateFormFile(fmt.Sprintf("file%d", i), file.Name())
		if err != nil {
			return err
		}

		if _, err := file.Transfer(part); err != nil {
			return err
		}
	}

	if err := mw.Close(); err != nil {
		return err
	}

	u := *c.base
	u.Path = "/api/ora/v1/upload"
	req, err := http.NewRequest(http.MethodPost, u.String(), &body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("x-scope", string(w.scopeID))
	req.Header.Set("x-receiver", id)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: w.opts.sessionID})

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}

	return nil
}

func (c *wsTransport) close(w *Window) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn == nil {
		return
	}

	close(c.done)
	_ = c.conn.Close()
	c.conn = nil
}

func (w *Window) connError() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.connErr
}
