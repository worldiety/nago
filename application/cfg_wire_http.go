// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"compress/gzip"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/core/http/poll"
	"go.wdy.de/nago/presentation/proto"
)

const (
	// wirePollWait is the longest time, which a poll of the http wire waits for messages. It stays below the idle
	// timeouts of common proxies, which are 30 to 60 seconds.
	wirePollWait = 25 * time.Second
	// wireChannelStale is the idle time, after which the scope of a channel may have expired, see wireKeepAlive. The
	// client gets a new channel then, which makes it render the window again.
	wireChannelStale = 45 * time.Second
	// wireMaxBody limits the messages of a request of the client.
	wireMaxBody = 64 * 1024 * 1024
	// wireChannelHeader carries the token of the channel, see [poll.Channel.Token].
	wireChannelHeader = "X-Nago-Channel"
)

// httpWire is the fallback of the websocket of /wire over plain http, for proxies and firewalls which block
// websockets. A client polls POST /wire/poll?_sid=<scope>&_c=<token>&ack=<n> for the messages of the scope, which
// acknowledges the messages up to n, and posts its own messages to POST /wire/send?_sid=<scope>&_c=<token>. Both
// bodies are frames, see [poll.WriteFrames]. A poll with an unknown token connects the scope to a new channel and
// answers with its token in the X-Nago-Channel header. A request for a closed or replaced channel fails with 409
// Conflict, then the client polls without a token.
type httpWire struct {
	cfg *Configurator
	app *core.Application

	mutex     sync.Mutex
	entries   map[proto.ScopeID]*wireEntry
	cleanedAt time.Time
}

type wireEntry struct {
	channel *poll.Channel
	scope   *core.Scope
	// session is the session, which owns the scope, see core.Application.ConnectSession. The token of the channel is
	// part of the address of a request, thus a request must also carry the cookie of this session.
	session session.ID
}

func newHTTPWire(cfg *Configurator, app *core.Application) *httpWire {
	return &httpWire{cfg: cfg, app: app, entries: map[proto.ScopeID]*wireEntry{}, cleanedAt: time.Now()}
}

// serve handles the requests below /wire, which are no websocket handshake, and tells whether it has handled r.
func (h *httpWire) serve(w http.ResponseWriter, r *http.Request) bool {
	var handle func(w http.ResponseWriter, r *http.Request, sid proto.ScopeID)
	switch r.URL.Path {
	case "/wire/poll":
		handle = h.poll
	case "/wire/send":
		handle = h.send
	default:
		return false
	}

	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return true
	}

	if !h.cfg.checkWireOrigin(r) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return true
	}

	// a scope keeps its id, unlike the websocket, which replaces a short one
	sid := proto.ScopeID(r.URL.Query().Get("_sid"))
	if len(sid) < 32 {
		http.Error(w, "invalid scope id", http.StatusBadRequest)
		return true
	}

	handle(w, r, sid)
	return true
}

func (h *httpWire) poll(w http.ResponseWriter, r *http.Request, sid proto.ScopeID) {
	query := r.URL.Query()
	var ack uint64
	if s := query.Get("ack"); s != "" {
		var err error
		if ack, err = strconv.ParseUint(s, 10, 64); err != nil {
			http.Error(w, "invalid ack", http.StatusBadRequest)
			return
		}
	}

	entry, cookies, ok := h.attach(r, sid, query.Get("_c"))
	if !ok {
		h.refuse(w, r, sid, cookies)
		return
	}

	wait := wirePollWait
	if cookies != nil {
		// a new channel answers at once, which tells the client its token
		ack = 0
		wait = 0
		for _, cookie := range cookies {
			http.SetCookie(w, cookie)
		}
	}

	entry.scope.Tick()
	first, msgs, err := entry.channel.Poll(r.Context(), ack, wait)
	if err != nil {
		if !errors.Is(err, poll.ErrClosed) {
			slog.Info("http wire poll failed", "err", err, "id", logging.Secret(string(sid)))
			h.drop(sid, entry)
		}

		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	entry.scope.Tick()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set(wireChannelHeader, entry.channel.Token())

	var out io.Writer = w
	if len(msgs) > 0 && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		out = gz
	}

	if err := poll.WriteFrames(out, first, msgs); err != nil {
		slog.Info("http wire poll cannot respond", "err", err, "id", logging.Secret(string(sid)))
	}
}

func (h *httpWire) send(w http.ResponseWriter, r *http.Request, sid proto.ScopeID) {
	entry := h.lookup(sid, r.URL.Query().Get("_c"))
	if entry == nil {
		http.Error(w, "unknown channel", http.StatusConflict)
		return
	}

	if requester, _ := h.cfg.sessionIDOf(r); session.ID(requester) != entry.session {
		slog.Warn("refused a send of another session to a channel", "id", logging.Secret(string(sid)))
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	first, msgs, err := poll.ReadFrames(http.MaxBytesReader(w, r.Body, wireMaxBody), wireMaxBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entry.scope.Tick()
	if err := entry.channel.Receive(first, msgs); err != nil {
		// like a websocket, whose loop fails: the client connects through a new channel
		slog.Info("http wire send failed", "err", err, "id", logging.Secret(string(sid)))
		h.drop(sid, entry)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// attach returns the channel of the scope with the token. Otherwise, it connects the scope to a new channel and
// returns the cookies of the session to set, which are not nil then. A scope of another session is refused with
// false and the cookies of the requesting session, and its channel stays as it is, see
// core.Application.ConnectSession: the scope id and the token are part of the address of a request, thus neither
// grants the window without the cookie of its session.
func (h *httpWire) attach(r *http.Request, sid proto.ScopeID, token string) (*wireEntry, []*http.Cookie, bool) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.cleanLocked()
	// the session comes from the http-only cookie, which the client cannot override over the wire
	requester, cookies := h.cfg.wireSession(r)
	entry := h.entries[sid]
	if entry != nil && entry.channel.Token() == token && !entry.channel.Closed() && entry.channel.Idle() < wireChannelStale {
		if entry.session != session.ID(requester) {
			slog.Warn("refused a poll of another session to a channel", "id", logging.Secret(string(sid)))
			return nil, cookies, false
		}

		return entry, nil, true
	}

	channel := poll.New()
	scope, ok := h.app.ConnectSession(channel, sid, session.ID(requester))
	if !ok {
		return nil, cookies, false
	}

	// the new channel replaced the former one at the scope already
	if entry != nil {
		entry.channel.Close()
		entry.scope.Disconnect(entry.channel)
	}

	entry = &wireEntry{channel: channel, scope: scope, session: session.ID(requester)}
	h.entries[sid] = entry
	if cookies == nil {
		cookies = []*http.Cookie{}
	}

	return entry, cookies, true
}

// refuse answers a poll, which must not connect to the scope, with a channel of its own, which only asks the page
// to load again, see core.RequestReload: the browser then gets a window of its own, and nothing of the refused
// window is sent.
func (h *httpWire) refuse(w http.ResponseWriter, r *http.Request, sid proto.ScopeID, cookies []*http.Cookie) {
	channel := poll.New()
	defer channel.Close()

	if err := core.RequestReload(channel); err != nil {
		slog.Info("cannot request a reload of a refused poll", "err", err, "id", logging.Secret(string(sid)))
	}

	first, msgs, err := channel.Poll(r.Context(), 0, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	for _, cookie := range cookies {
		http.SetCookie(w, cookie)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set(wireChannelHeader, channel.Token())
	if err := poll.WriteFrames(w, first, msgs); err != nil {
		slog.Info("http wire cannot respond to a refused poll", "err", err, "id", logging.Secret(string(sid)))
	}
}

func (h *httpWire) lookup(sid proto.ScopeID, token string) *wireEntry {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	entry := h.entries[sid]
	if entry == nil || entry.channel.Token() != token || entry.channel.Closed() {
		return nil
	}

	return entry
}

// drop closes the channel and detaches it from its scope, unless the scope has another channel already.
func (h *httpWire) drop(sid proto.ScopeID, entry *wireEntry) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.entries[sid] == entry {
		delete(h.entries, sid)
	}

	entry.channel.Close()
	entry.scope.Disconnect(entry.channel)
}

// cleanLocked removes the channels, whose client has gone, from time to time.
func (h *httpWire) cleanLocked() {
	if time.Since(h.cleanedAt) < wireChannelStale {
		return
	}

	h.cleanedAt = time.Now()
	for sid, entry := range h.entries {
		if entry.channel.Closed() || entry.channel.Idle() > 2*wireChannelStale {
			delete(h.entries, sid)
			entry.channel.Close()
			entry.scope.Disconnect(entry.channel)
		}
	}
}

// wireSession returns the session of a request of the wire and the cookies to set. The page normally got its cookie
// already. If not, e.g. because all assets came from the browser cache after the page has been reached by a link from
// another site, the wire issues it: the wire is always called by the page itself, so the cookie cannot be missing just
// due to SameSite. A page in a frame of another site has the embedded cookie instead, see
// [Configurator.EmbeddablePages], thus the wire issues none: a browser would drop it there.
func (c *Configurator) wireSession(r *http.Request) (sessionID string, cookies []*http.Cookie) {
	sessionID, migrate := c.sessionIDOf(r)
	switch {
	case migrate:
		return sessionID, []*http.Cookie{c.newSessionCookie(r, sessionID), expiredPlainSessionCookie()}
	case sessionID == "":
		cookie := c.newSessionCookie(r, string(proto.NewScopeID()))
		return cookie.Value, []*http.Cookie{cookie}
	}

	return sessionID, nil
}
