// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/core/http/poll"
	"go.wdy.de/nago/presentation/proto"
)

// DialHTTP connects a new window like [Dial], but through the http fallback of the wire, which the frontend uses,
// when a proxy or firewall blocks the websocket.
func DialHTTP(t TB, baseURL string, path core.NavigationPath, opts ...OpenOption) *Window {
	t.Helper()

	base, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("nagotest: invalid url %q: %v", baseURL, err)
	}

	w := newWindow(t, &httpTransport{base: base}, path, opts)
	w.open()
	return w
}

// httpTransport connects a window through the http fallback of the wire: it polls the messages of the scope and
// posts its own ones, just like the frontend.
type httpTransport struct {
	base   *url.URL
	client http.Client

	mutex   sync.Mutex // orders the sends and guards the fields below
	token   string
	sendSeq uint64
	cancel  context.CancelFunc
	done    chan struct{}
}

func (c *httpTransport) request(ctx context.Context, w *Window, path string, query url.Values, body []byte) (*http.Response, error) {
	u := *c.base
	u.Path = path
	query.Set("_sid", string(w.scopeID))
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	for _, name := range w.opts.cookieNames() {
		req.AddCookie(&http.Cookie{Name: name, Value: w.opts.sessionID})
	}

	return c.client.Do(req)
}

func (c *httpTransport) connect(w *Window) error {
	// a poll without a token connects the scope to a new channel and answers at once
	resp, err := c.request(context.Background(), w, "/wire/poll", url.Values{}, nil)
	if err != nil {
		return err
	}

	token := resp.Header.Get("X-Nago-Channel")
	ack, err := c.receive(w, resp, 0)
	if err != nil {
		return err
	}

	if token == "" {
		return errors.New("the poll has not answered with a channel")
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	c.mutex.Lock()
	c.token = token
	c.sendSeq = 0
	c.cancel = cancel
	c.done = done
	c.mutex.Unlock()

	go c.poll(ctx, w, token, ack, done)
	return nil
}

func (c *httpTransport) poll(ctx context.Context, w *Window, token string, ack uint64, done chan struct{}) {
	defer close(done)
	for {
		resp, err := c.request(ctx, w, "/wire/poll", url.Values{"_c": {token}, "ack": {strconv.FormatUint(ack, 10)}}, nil)
		if err == nil {
			ack, err = c.receive(w, resp, ack)
		}

		if err != nil {
			if ctx.Err() == nil {
				w.fail(fmt.Errorf("connection lost: %w", err))
			}

			return
		}
	}
}

// receive delivers the messages of a poll, which follow ack, and returns the new ack.
func (c *httpTransport) receive(w *Window, resp *http.Response, ack uint64) (uint64, error) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return ack, fmt.Errorf("poll failed with status %d: %s", resp.StatusCode, bytes.TrimSpace(msg))
	}

	first, msgs, err := poll.ReadFrames(resp.Body, 1<<30)
	if err != nil {
		return ack, err
	}

	for i, msg := range msgs {
		seq := first + uint64(i)
		if seq <= ack {
			continue
		}

		obj, err := proto.Unmarshal(proto.NewBinaryReader(bytes.NewBuffer(msg)))
		if err != nil {
			return ack, fmt.Errorf("cannot decode message: %w", err)
		}

		evt, ok := obj.(proto.NagoEvent)
		if !ok {
			return ack, unexpected(obj)
		}

		w.receive(evt)
		ack = seq
	}

	return ack, nil
}

func (c *httpTransport) write(w *Window, evt proto.NagoEvent) error {
	var msg bytes.Buffer
	if err := proto.Marshal(proto.NewBinaryWriter(&msg), evt); err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.token == "" {
		return errors.New("not connected")
	}

	var body bytes.Buffer
	if err := poll.WriteFrames(&body, c.sendSeq+1, [][]byte{msg.Bytes()}); err != nil {
		return err
	}

	resp, err := c.request(context.Background(), w, "/wire/send", url.Values{"_c": {c.token}}, body.Bytes())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("send failed with status %d: %s", resp.StatusCode, bytes.TrimSpace(msg))
	}

	c.sendSeq++
	return nil
}

func (c *httpTransport) send(w *Window, evt proto.NagoEvent) error {
	return c.write(w, evt)
}

func (c *httpTransport) settle(w *Window) error {
	return settle(w, func(evt proto.NagoEvent) error { return c.write(w, evt) })
}

func (c *httpTransport) upload(w *Window, id string, files []core.File) error {
	return upload(&c.client, c.base, w, id, files)
}

func (c *httpTransport) close(w *Window) {
	c.mutex.Lock()
	cancel, done := c.cancel, c.done
	c.token = ""
	c.cancel = nil
	c.done = nil
	c.mutex.Unlock()

	if cancel != nil {
		cancel()
		<-done
	}
}
