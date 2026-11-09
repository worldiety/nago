// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nethost_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/lib/net/nethost"
	"go.wdy.de/nago/pkg/gostrict/std/task"
	"go.wdy.de/nago/pkg/gostrict/std/task/taskhost"
)

func TestMatch(t *testing.T) {
	declared := func(method, path string) (string, bool) {
		if method == "GET" && strings.HasPrefix(path, "/customers/") {
			return "/customers/{id}", true
		}
		return "", false
	}
	for _, tt := range []struct {
		name      string
		match     func(method, path string) (string, bool)
		method    string
		path      string
		end       string
		hits      int32
		decorated int32
		route     string
		seen      string
	}{
		{name: "declared", match: declared, method: "GET", path: "/customers/17", end: "ok|200||hello|", hits: 1, decorated: 1, route: "/customers/{id}", seen: "GET /customers/17"},
		{name: "an empty method is GET", match: declared, method: "", path: "/customers/17", end: "ok|200||hello|", hits: 1, decorated: 1, route: "/customers/{id}", seen: "GET /customers/17"},
		{name: "the plain path before its encoding", match: declared, method: "GET", path: "/customers/a b", end: "ok|200||hello|", hits: 1, decorated: 1, route: "/customers/{id}", seen: "GET /customers/a b"},
		{name: "another method is not declared", match: declared, method: "POST", path: "/customers/17", end: "failed|0|the endpoint is not declared||", seen: "POST /customers/17"},
		{name: "another path is not declared", match: declared, method: "GET", path: "/orders/1", end: "failed|0|the endpoint is not declared||", seen: "GET /orders/1"},
		{name: "no match allows everything without a route", match: nil, method: "POST", path: "/orders/1", end: "ok|200||hello|", hits: 1, decorated: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var hits, decorated atomic.Int32
			srv := httptest.NewServer(countingHandler(&hits, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "hello") }))
			t.Cleanup(srv.Close)
			var seen atomic.Value
			r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true}, conn: func(c *nethost.Connection) {
				c.Decorate = func(next http.RoundTripper) http.RoundTripper {
					decorated.Add(1)
					return next
				}
				if tt.match != nil {
					c.Match = func(method, path string) (string, bool) {
						seen.Store(method + " " + path)
						return tt.match(method, path)
					}
				}
			}})
			r.send("Send", tt.method, tt.path)
			if got := r.end(); got != tt.end {
				t.Errorf("end: got %q, want %q", got, tt.end)
			}
			if hits.Load() != tt.hits || decorated.Load() != tt.decorated {
				t.Errorf("hits %d decorated %d, want %d and %d", hits.Load(), decorated.Load(), tt.hits, tt.decorated)
			}
			if got, _ := seen.Load().(string); got != tt.seen {
				t.Errorf("match saw %q, want %q", got, tt.seen)
			}
			if obs := r.observed(); len(obs) != 1 || obs[0].Route != tt.route {
				t.Errorf("observations: %+v, want one of the route %q", obs, tt.route)
			}
		})
	}
}

func TestResolverErrors(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		end  string
	}{
		{name: "refused", err: nethost.Refused("the secret belongs to another server"), end: "failed|0|the secret belongs to another server||"},
		{name: "wrapped refused", err: fmt.Errorf("resolve: %w", nethost.Refused("not here")), end: "failed|0|not here||"},
		{name: "any other error stays hidden", err: errors.New("the database at 10.1.2.3 is down"), end: "failed|0|the connection is not available||"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := newRig(t, rigConfig{url: "http://127.0.0.1:1", resolve: func(string) (nethost.Connection, error) {
				return nethost.Connection{}, tt.err
			}})
			r.send("Send", "GET", "/x")
			if got := r.end(); got != tt.end {
				t.Errorf("got %q, want %q", got, tt.end)
			}
		})
	}
}

// TestAbortOfTheHost shows, that an abort of a function of the host reaches the program as an abort, while another panic
// only ends the request, and that neither leaves the request open.
func TestAbortOfTheHost(t *testing.T) {
	aborted := errors.New("the trial run may not send")
	abort := func() { gostrict.Abort(aborted) }
	defect := func() { panic("defect") }
	for _, tt := range []struct {
		name   string
		cfg    func(fn func()) rigConfig
		fn     func()
		abort  bool
		end    string
		method string
	}{
		{name: "abort in the resolver", abort: true, fn: abort, cfg: func(fn func()) rigConfig {
			return rigConfig{resolve: func(string) (nethost.Connection, error) { fn(); return nethost.Connection{}, nil }}
		}},
		{name: "defect in the resolver", fn: defect, end: "failed|0|the connection is not available||", cfg: func(fn func()) rigConfig {
			return rigConfig{resolve: func(string) (nethost.Connection, error) { fn(); return nethost.Connection{}, nil }}
		}},
		{name: "abort in the streams", abort: true, fn: abort, method: "Download", cfg: func(fn func()) rigConfig {
			return rigConfig{streams: panicStreams{fn}}
		}},
		{name: "defect in the streams", fn: defect, method: "Download", end: "failed|0|the file is not available||", cfg: func(fn func()) rigConfig {
			return rigConfig{streams: panicStreams{fn}}
		}},
		{name: "abort in defer", abort: true, fn: abort, cfg: func(fn func()) rigConfig {
			return rigConfig{defers: func(_, _ func()) { fn() }}
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var hits atomic.Int32
			srv := httptest.NewServer(countingHandler(&hits, func(http.ResponseWriter, *http.Request) {}))
			t.Cleanup(srv.Close)
			q := newQueue()
			jobs := taskhost.New(q.post)
			cfg := tt.cfg(tt.fn)
			cfg.url, cfg.jobs = srv.URL, jobs
			cfg.opts.AllowPrivateNetworks = true
			r := newRig(t, cfg)

			var res []any
			var err error
			if tt.method == "Download" {
				res, err = r.pool.Call(t.Context(), "Download", "/x", "file")
			} else {
				res, err = r.pool.Call(t.Context(), "Send", "GET", "/x")
			}
			if tt.abort {
				if !errors.Is(err, aborted) {
					t.Fatalf("the program must be aborted: %v", err)
				}
				q.next(t) // the job has ended
				if jobs.Len() != 0 {
					t.Error("the job is still open")
				}
				if len(r.ends) != 0 || r.doneCalls() != 0 {
					t.Error("done must not run after an abort")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if got := r.end(); got != tt.end {
					t.Errorf("got %q, want %q", got, tt.end)
				}
				jobs.Wait(task.ID(res[0].(string)), func(string) {})
				q.next(t)
			}
			r.waitOpen(0)
			if hits.Load() != 0 {
				t.Error("nothing must be sent")
			}
		})
	}
}

type panicStreams struct{ fn func() }

func (s panicStreams) Source(string) (nethost.Source, error) { s.fn(); return nethost.Source{}, nil }
func (s panicStreams) Sink(string) (nethost.Sink, error)     { s.fn(); return nil, nil }

func TestPrivateOnlyRule(t *testing.T) {
	for _, tt := range []struct {
		ip      string
		private bool
		want    bool
	}{
		{"127.0.0.1", true, false},
		{"10.0.0.1", true, false},
		{"192.168.1.1", true, false},
		{"100.64.0.1", true, false},
		{"fd12::1", true, false},
		{"::1", true, false},
		// the same addresses without the release of the private networks
		{"127.0.0.1", false, true},
		{"10.0.0.1", false, true},
		{"100.64.0.1", false, true},
		// public addresses are what the rule is there for: the same address passes the normal rule, see below
		{"8.8.8.8", true, true},
		{"2001:4860:4860::8888", true, true},
		// the rules of all connections stay
		{"169.254.169.254", true, true},
		{"100.100.100.200", true, true},
	} {
		if got := nethost.ForbiddenPrivateOnly(netip.MustParseAddr(tt.ip), tt.private); got != tt.want {
			t.Errorf("%s private=%v: got %v", tt.ip, tt.private, got)
		}
	}
	if nethost.ForbiddenAddress(netip.MustParseAddr("8.8.8.8"), true) {
		t.Error("a public address must pass the normal rule")
	}
}

func TestPrivateOnly(t *testing.T) {
	for _, tt := range []struct {
		name        string
		allow, only bool
		end         string
		class       nethost.Class
		hits        int32
	}{
		{name: "private network released", allow: true, only: true, end: "ok|200||hello|", class: nethost.Answered, hits: 1},
		{name: "private network not released", allow: false, only: true, end: "failed|0|the server can not be reached||", class: nethost.NotSent},
		{name: "without the rule and released", allow: true, only: false, end: "ok|200||hello|", class: nethost.Answered, hits: 1},
		{name: "without the rule and not released", allow: false, only: false, end: "failed|0|the server can not be reached||", class: nethost.NotSent},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var hits atomic.Int32
			srv := httptest.NewServer(countingHandler(&hits, func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "hello") }))
			t.Cleanup(srv.Close)
			r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: tt.allow}, conn: func(c *nethost.Connection) { c.PrivateOnly = tt.only }})
			r.send("Send", "GET", "/x")
			if got := r.end(); got != tt.end {
				t.Errorf("end: got %q, want %q", got, tt.end)
			}
			if hits.Load() != tt.hits {
				t.Errorf("hits: got %d, want %d", hits.Load(), tt.hits)
			}
			if obs := r.observed(); len(obs) != 1 || obs[0].Class != tt.class {
				t.Errorf("observations: %+v", obs)
			}
		})
	}
}

// TestPrivateOnlyNeverReusesAConnection counts the connections of two requests in a row: a connection is only reused
// by a request of the same rule.
func TestPrivateOnlyNeverReusesAConnection(t *testing.T) {
	for _, tt := range []struct {
		name          string
		first, second bool
		connections   int32
	}{
		{name: "both normal share a connection", first: false, second: false, connections: 1},
		{name: "both private only share a connection", first: true, second: true, connections: 1},
		{name: "a private only request can not use a normal connection", first: false, second: true, connections: 2},
		{name: "a normal request can not use a private only connection", first: true, second: false, connections: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var opened atomic.Int32
			srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "hello") }))
			srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
				if s == http.StateNew {
					opened.Add(1)
				}
			}
			srv.Start()
			t.Cleanup(srv.Close)
			var next atomic.Bool
			next.Store(tt.first)
			r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true}, conn: func(c *nethost.Connection) { c.PrivateOnly = next.Load() }})
			r.send("Send", "GET", "/1")
			if got := r.end(); got != "ok|200||hello|" {
				t.Fatal(got)
			}
			next.Store(tt.second)
			r.send("Send", "GET", "/2")
			if got := r.end(); got != "ok|200||hello|" {
				t.Fatal(got)
			}
			if opened.Load() != tt.connections {
				t.Errorf("connections: got %d, want %d", opened.Load(), tt.connections)
			}
		})
	}
}

// TestObservations is the table of the ends of a request. A row, which is not sent, also shows that the server was
// not reached, and its neighbours show that the server would have answered.
func TestObservations(t *testing.T) {
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api/ok", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "hello") })
	mux.HandleFunc("/api/nf", func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "no", http.StatusNotFound) })
	mux.HandleFunc("/api/boom", func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "boom", http.StatusInternalServerError) })
	mux.HandleFunc("/api/big", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, strings.Repeat("x", 100)) })
	mux.HandleFunc("/api/hop", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/api/ok", http.StatusFound) })
	mux.HandleFunc("/api/away", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://elsewhere.example/", http.StatusFound)
	})
	mux.HandleFunc("/api/cut", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "1000")
		fmt.Fprint(w, "half")
		w.(http.Flusher).Flush()
		panic(http.ErrAbortHandler)
	})
	mux.HandleFunc("/api/hang", func(_ http.ResponseWriter, r *http.Request) { <-r.Context().Done() })
	mux.HandleFunc("/api/hang-body", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "half")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	})
	srv := httptest.NewServer(countingHandler(&hits, mux.ServeHTTP))
	t.Cleanup(srv.Close)
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()

	match := func(_, p string) (string, bool) {
		if p == "/undeclared" {
			return "", false
		}
		return "route", true
	}
	sinks := func() *fakeStreams {
		return &fakeStreams{
			sources: map[string]nethost.Source{"unopenable": {Size: 3, Open: func() (io.ReadCloser, error) { return nil, errors.New("gone") }}},
			sinks: map[string]nethost.Sink{
				"broken":     &memSink{writeErr: errors.New("disk full")},
				"quota":      &memSink{writeErr: nethost.Refused("the answer is too large")},
				"uncommited": &memSink{commitErr: errors.New("disk full")},
			},
		}
	}

	for _, tt := range []struct {
		name     string
		fn       string
		args     []any
		url      string
		blocked  bool
		maxBody  int64
		refused  error
		decorate func(http.RoundTripper) http.RoundTripper
		end      string
		class    nethost.Class
		status   int32
		method   string
		route    string
		sentZero bool
	}{
		{name: "complete answer", fn: "Send", args: []any{"GET", "/ok"}, end: "ok|200||hello|", class: nethost.Answered, status: 200, method: "GET", route: "route"},
		{name: "empty method", fn: "Send", args: []any{"", "/ok"}, end: "ok|200||hello|", class: nethost.Answered, status: 200, method: "GET", route: "route"},
		{name: "client error", fn: "Send", args: []any{"GET", "/nf"}, end: "rejected|404|HTTP 404: no\n||", class: nethost.Answered, status: 404, method: "GET", route: "route"},
		{name: "server error", fn: "Send", args: []any{"GET", "/boom"}, end: "rejected|500|HTTP 500: boom\n||", class: nethost.Answered, status: 500, method: "GET", route: "route"},
		{name: "text over the limit", fn: "Send", args: []any{"GET", "/big"}, maxBody: 10, end: "unknown|200|the answer is too large||", class: nethost.Answered, status: 200, method: "GET", route: "route"},
		{name: "text at the limit", fn: "Send", args: []any{"GET", "/big"}, maxBody: 100, end: "ok|200||" + strings.Repeat("x", 100) + "|", class: nethost.Answered, status: 200, method: "GET", route: "route"},
		{name: "redirect in the origin counts once", fn: "Send", args: []any{"GET", "/hop"}, end: "ok|200||hello|", class: nethost.Answered, status: 200, method: "GET", route: "route"},
		{name: "redirect to another origin is the last status", fn: "Send", args: []any{"GET", "/away"}, class: nethost.Answered, status: 302, method: "GET", route: "route"},
		{name: "sink not writable", fn: "Download", args: []any{"/ok", "broken"}, end: "unknown|200|the answer can not be stored||", class: nethost.Answered, status: 200, method: "GET", route: "route"},
		{name: "sink refuses", fn: "Download", args: []any{"/ok", "quota"}, end: "unknown|200|the answer is too large||", class: nethost.Answered, status: 200, method: "GET", route: "route"},
		{name: "commit fails", fn: "Download", args: []any{"/ok", "uncommited"}, end: "unknown|200|the answer can not be stored||", class: nethost.Answered, status: 200, method: "GET", route: "route"},
		{name: "status and then the connection is lost", fn: "Send", args: []any{"GET", "/cut"}, end: "unknown|200|the connection was lost after the request was sent||", class: nethost.Network, method: "GET", route: "route"},
		{name: "status and then the time ends", fn: "Timed", args: []any{"/hang-body", int32(30)}, end: "unknown|200|the request has timed out after it was sent||", class: nethost.Network, method: "GET", route: "route"},
		{name: "no answer in time", fn: "Timed", args: []any{"/hang", int32(30)}, end: "unknown|0|the request has timed out after it was sent||", class: nethost.Network, method: "GET", route: "route"},
		{name: "server not reachable", fn: "Send", args: []any{"GET", "/ok"}, url: closed.URL, end: "failed|0|the server can not be reached||", class: nethost.Network, method: "GET", route: "route", sentZero: true},
		{name: "forbidden address", fn: "Send", args: []any{"GET", "/ok"}, blocked: true, end: "failed|0|the server can not be reached||", class: nethost.NotSent, method: "GET", route: "route", sentZero: true},
		{name: "not declared", fn: "Send", args: []any{"GET", "/undeclared"}, end: "failed|0|the endpoint is not declared||", class: nethost.NotSent, method: "GET", sentZero: true},
		{name: "invalid path", fn: "Send", args: []any{"GET", "x"}, end: "failed|0|the path must begin with /||", class: nethost.NotSent, method: "GET", sentZero: true},
		{name: "invalid method is not recorded", fn: "Send", args: []any{"Kundin Erika", "/ok"}, end: "failed|0|the method is not allowed||", class: nethost.NotSent, sentZero: true},
		{name: "forbidden header keeps the route", fn: "Header", args: []any{"/ok", "Cookie", "x"}, end: "failed|0|the header Cookie is not allowed||", class: nethost.NotSent, method: "GET", route: "route", sentZero: true},
		{name: "invalid value of a header", fn: "Header", args: []any{"/ok", "X-A", "a\nb"}, end: "failed|0|the header X-A is not allowed||", class: nethost.NotSent, method: "GET", route: "route", sentZero: true},
		{name: "invalid name of a header", fn: "Header", args: []any{"/ok", "X A", "a"}, end: "failed|0|the header X A is not allowed||", class: nethost.NotSent, method: "GET", route: "route", sentZero: true},
		{name: "a defect before the sending", fn: "Send", args: []any{"GET", "/ok"}, end: "unknown|0|the request has failed||", class: nethost.NotSent, method: "GET", route: "route", sentZero: true,
			decorate: func(http.RoundTripper) http.RoundTripper {
				return roundTripper(func(*http.Request) (*http.Response, error) { panic("defect") })
			}},
		{name: "refused connection", fn: "Send", args: []any{"GET", "/ok"}, refused: nethost.Refused("another server"), end: "failed|0|another server||", class: nethost.NotSent, method: "GET", sentZero: true},
		{name: "the source can not be opened", fn: "Upload", args: []any{"/ok", "unopenable"}, end: "failed|0|the file can not be read||", class: nethost.NotSent, method: "POST", route: "route", sentZero: true},
		{name: "the source does not exist", fn: "Upload", args: []any{"/ok", "unknown"}, end: "failed|0|the file is not available||", class: nethost.NotSent, method: "POST", route: "route", sentZero: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			hits.Store(0)
			base := srv.URL
			if tt.url != "" {
				base = tt.url
			}
			parsed, err := url.Parse(base + "/api")
			if err != nil {
				t.Fatal(err)
			}
			begin := time.Now()
			r := newRig(t, rigConfig{
				url:     base,
				streams: sinks(),
				opts:    nethost.Options{AllowPrivateNetworks: !tt.blocked, MaxBody: tt.maxBody},
				resolve: func(string) (nethost.Connection, error) {
					if tt.refused != nil {
						return nethost.Connection{}, tt.refused
					}
					return nethost.Connection{BaseURL: parsed, Match: match, Decorate: tt.decorate}, nil
				},
			})
			r.send(tt.fn, tt.args...)
			got := r.end()
			if tt.end != "" && got != tt.end {
				t.Errorf("end: got %q, want %q", got, tt.end)
			}
			if tt.end == "" && !strings.HasPrefix(got, "rejected|302|") {
				t.Errorf("end: got %q", got)
			}
			if tt.sentZero && hits.Load() != 0 {
				t.Errorf("the request reached the server: %d", hits.Load())
			}
			obs := r.observed()
			if len(obs) != 1 {
				t.Fatalf("got %d observations: %+v", len(obs), obs)
			}
			o := obs[0]
			if o.Connection != "api" || o.Method != tt.method || o.Route != tt.route || o.Class != tt.class || o.Status != tt.status {
				t.Errorf("observation %+v, want class %d status %d method %q route %q", o, tt.class, tt.status, tt.method, tt.route)
			}
			if o.End.Before(begin) || o.End.After(time.Now()) || o.Duration < 0 {
				t.Errorf("implausible time: %+v", o)
			}
			if seen := r.seenAtDone; len(seen) != 1 || seen[0] != 1 {
				t.Errorf("the observation must exist before done: %v", seen)
			}
		})
	}
}

// TestObservationTimes pins the time of an observation between what the server and the sink have seen.
func TestObservationTimes(t *testing.T) {
	var firstHop atomic.Pointer[time.Time]
	var lastByte atomic.Pointer[time.Time]
	mux := http.NewServeMux()
	mux.HandleFunc("/api/hop", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		firstHop.Store(&now)
		http.Redirect(w, r, "/api/file", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/api/file", func(w http.ResponseWriter, _ *http.Request) {
		now := time.Now()
		lastByte.Store(&now)
		fmt.Fprint(w, "content")
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	var commitBegan atomic.Pointer[time.Time]
	sink := &memSink{onCommit: func() {
		now := time.Now()
		commitBegan.Store(&now)
	}}
	r := newRig(t, rigConfig{url: srv.URL, opts: nethost.Options{AllowPrivateNetworks: true}, streams: &fakeStreams{sinks: map[string]nethost.Sink{"s": sink}}})
	r.send("Download", "/hop", "s")
	if got := r.end(); got != "ok|200|||" {
		t.Fatal(got)
	}
	obs := r.observed()
	if len(obs) != 1 {
		t.Fatalf("a redirect is one observation: %+v", obs)
	}
	began := obs[0].End.Add(-obs[0].Duration)
	if !began.Before(*firstHop.Load()) {
		t.Errorf("the duration must begin before the first step: began %v, first step %v", began, *firstHop.Load())
	}
	if !obs[0].End.After(*lastByte.Load()) {
		t.Errorf("the end must be after the last step: %v, %v", obs[0].End, *lastByte.Load())
	}
	if obs[0].End.After(*commitBegan.Load()) {
		t.Errorf("the end is the last byte, not the commit: %v, %v", obs[0].End, *commitBegan.Load())
	}
}

// TestConnectionContext ends the context of the connection at every stage of a request, next to a request, whose
// context stays alive.
func TestConnectionContext(t *testing.T) {
	type stage int
	const (
		alive stage = iota
		beforeStart
		waitingForAnswer
		duringTransfer
		duringCommit
	)
	for _, tt := range []struct {
		name    string
		stage   stage
		hits    int32
		done    bool
		commits int
		aborts  int
	}{
		{name: "alive", stage: alive, hits: 1, done: true, commits: 1},
		{name: "before the start", stage: beforeStart, hits: 0, aborts: 1},
		{name: "waiting for the answer", stage: waitingForAnswer, hits: 1, aborts: 1},
		{name: "during the transfer", stage: duringTransfer, hits: 1, aborts: 1},
		{name: "after the complete answer", stage: duringCommit, hits: 1, commits: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			var hits atomic.Int32
			entered := make(chan struct{}, 1)
			srv := httptest.NewServer(countingHandler(&hits, func(w http.ResponseWriter, r *http.Request) {
				switch tt.stage {
				case waitingForAnswer:
					entered <- struct{}{}
					<-r.Context().Done()
				case duringTransfer:
					fmt.Fprint(w, "half")
					w.(http.Flusher).Flush()
					<-r.Context().Done()
				default:
					fmt.Fprint(w, "content")
				}
			}))
			t.Cleanup(srv.Close)

			wrote := make(chan struct{}, 16)
			sink := &memSink{onWrite: func() { wrote <- struct{}{} }}
			if tt.stage == duringCommit {
				sink.onCommit = cancel
			}
			var start func()
			q := newQueue()
			jobs := taskhost.New(q.post)
			r := newRig(t, rigConfig{
				url: srv.URL, jobs: jobs,
				opts:    nethost.Options{AllowPrivateNetworks: true},
				streams: &fakeStreams{sinks: map[string]nethost.Sink{"s": sink}},
				conn:    func(c *nethost.Connection) { c.Context = ctx },
				defers:  func(s, _ func()) { start = s },
			})
			id := r.send("Download", "/x", "s")
			if tt.stage == beforeStart {
				cancel()
			}
			start()

			switch tt.stage {
			case waitingForAnswer:
				<-entered
				cancel()
			case duringTransfer:
				<-wrote
				cancel()
			}
			failure := jobFailure(t, q, jobs, id)

			commits, aborts, _ := sink.state()
			if commits != tt.commits || aborts != tt.aborts || hits.Load() != tt.hits {
				t.Errorf("commits %d aborts %d hits %d, want %d %d %d", commits, aborts, hits.Load(), tt.commits, tt.aborts, tt.hits)
			}
			if tt.done {
				if got := r.end(); got != "ok|200|||" {
					t.Errorf("end: %q", got)
				}
				if failure != "ok|200|||" {
					t.Errorf("failure: %q", failure)
				}
				if len(r.observed()) != 1 {
					t.Error("the observation is missing")
				}
			} else {
				if failure != "the request was cancelled" {
					t.Errorf("failure: %q", failure)
				}
				if len(r.ends) != 0 || r.doneCalls() != 0 {
					t.Error("done must not run")
				}
				if obs := r.observed(); len(obs) != 0 {
					t.Errorf("no observation expected: %+v", obs)
				}
			}
			r.waitOpen(0)
		})
	}
}

// TestDefer decides each request in every way a binding can, and for a refused request as well.
func TestDefer(t *testing.T) {
	type decision struct {
		name    string
		act     func(start, drop func())
		started bool
	}
	decisions := []decision{
		{"start", func(start, _ func()) { start() }, true},
		{"drop", func(_, drop func()) { drop() }, false},
		{"start and then drop", func(start, drop func()) { start(); drop() }, true},
		{"drop and then start", func(start, drop func()) { drop(); start() }, false},
		{"start twice", func(start, _ func()) { start(); start() }, true},
		{"drop twice", func(_, drop func()) { drop(); drop() }, false},
	}
	for _, kind := range []struct {
		name   string
		path   string
		end    string
		hits   int32
		commit int
	}{
		{name: "sendable", path: "/x", end: "ok|200|||", hits: 1, commit: 1},
		{name: "refused", path: "/undeclared", end: "failed|0|the endpoint is not declared||"},
	} {
		for _, d := range decisions {
			t.Run(kind.name+"/"+d.name, func(t *testing.T) {
				var hits atomic.Int32
				srv := httptest.NewServer(countingHandler(&hits, func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "content") }))
				t.Cleanup(srv.Close)

				var start, drop func()
				deferred := 0
				sink := &memSink{}
				q := newQueue()
				jobs := taskhost.New(q.post)
				r := newRig(t, rigConfig{
					url: srv.URL, jobs: jobs,
					opts:    nethost.Options{AllowPrivateNetworks: true},
					streams: &fakeStreams{sinks: map[string]nethost.Sink{"s": sink}},
					conn: func(c *nethost.Connection) {
						c.Match = func(_, p string) (string, bool) { return "r", p != "/undeclared" }
					},
					defers: func(s, dr func()) { deferred++; start, drop = s, dr },
				})
				id := r.send("Download", kind.path, "s")
				if deferred != 1 {
					t.Fatalf("defer was called %d times", deferred)
				}
				if hits.Load() != 0 || r.host.OpenRequests() != 1 {
					t.Fatalf("a deferred request must not be sent but must count: hits %d open %d", hits.Load(), r.host.OpenRequests())
				}

				d.act(start, drop)
				failure := jobFailure(t, q, jobs, id)

				commits, aborts, _ := sink.state()
				if d.started {
					if got := r.end(); got != kind.end {
						t.Errorf("end: got %q, want %q", got, kind.end)
					}
					if len(r.observed()) != 1 || hits.Load() != kind.hits || commits != kind.commit || failure != kind.end {
						t.Errorf("observations %d hits %d commits %d failure %q", len(r.observed()), hits.Load(), commits, failure)
					}
				} else {
					if len(r.ends) != 0 || r.doneCalls() != 0 || len(r.observed()) != 0 || hits.Load() != 0 || commits != 0 {
						t.Errorf("a drop must leave nothing behind: done %d observations %d hits %d commits %d", r.doneCalls(), len(r.observed()), hits.Load(), commits)
					}
					if failure != "the request was not sent" {
						t.Errorf("failure: %q", failure)
					}
					// only a sendable request has a sink, which a drop must discard
					if want := min(kind.commit, 1); aborts != want {
						t.Errorf("aborts: got %d, want %d", aborts, want)
					}
				}
				r.waitOpen(0)
				if len(q.fns) != 0 {
					t.Error("the job has ended twice")
				}
			})
		}
	}
}

// TestDeferCountsFromTheSend shows that a held request takes its place at the call of net.Send and that a drop returns
// it, for the limit of the host and for the limit of a window.
func TestDeferCountsFromTheSend(t *testing.T) {
	for _, tt := range []struct {
		name string
		opts nethost.Options
		jobs bool
	}{
		{name: "limit of the host", opts: nethost.Options{MaxOpen: 1, AllowPrivateNetworks: true}},
		{name: "limit of a window", opts: nethost.Options{MaxOpenPerWindow: 1, AllowPrivateNetworks: true}, jobs: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "hello") }))
			t.Cleanup(srv.Close)
			var starts, drops []func()
			q := newQueue()
			cfg := rigConfig{url: srv.URL, opts: tt.opts, defers: func(s, d func()) { starts, drops = append(starts, s), append(drops, d) }}
			if tt.jobs {
				cfg.jobs = taskhost.New(q.post)
			}
			r := newRig(t, cfg)

			r.send("Send", "GET", "/held")
			r.send("Send", "GET", "/refused")
			starts[1]()
			if got := r.end(); got != "failed|0|too many open requests||" {
				t.Errorf("the held request must count: %q", got)
			}
			drops[0]()
			if tt.jobs {
				q.next(t)
			}
			r.waitOpen(0)

			r.send("Send", "GET", "/free")
			starts[2]()
			if got := r.end(); got != "ok|200||hello|" {
				t.Errorf("a drop must free its place: %q", got)
			}
		})
	}
}

// TestForgottenDeferEndsTheRequest shows that a binding, which loses both functions, does not hold a place forever.
func TestForgottenDeferEndsTheRequest(t *testing.T) {
	q := newQueue()
	jobs := taskhost.New(q.post)
	sink := &memSink{}
	r := newRig(t, rigConfig{
		url: "http://127.0.0.1:1", jobs: jobs,
		opts:    nethost.Options{AllowPrivateNetworks: true},
		streams: &fakeStreams{sinks: map[string]nethost.Sink{"s": sink}},
		defers:  func(_, _ func()) {},
	})
	id := r.send("Download", "/x", "s")
	if r.host.OpenRequests() != 1 {
		t.Fatal("the request must count")
	}
	var failure string
	jobs.Wait(id, func(f string) { failure = f })
	deadline := time.Now().Add(20 * time.Second)
	for {
		runtime.GC()
		select {
		case fn := <-q.fns:
			fn()
			if failure != "the request was not sent" {
				t.Errorf("failure: %q", failure)
			}
			if _, aborts, _ := sink.state(); aborts != 1 {
				t.Errorf("aborts: %d", aborts)
			}
			r.waitOpen(0)
			return
		case <-time.After(10 * time.Millisecond):
		}
		if time.Now().After(deadline) {
			t.Fatal("the forgotten request was never ended")
		}
	}
}

// TestLimitsCountRefusedAtTheirEndpoint shows that a request, which the limits refuse, is observed and opens no sink.
func TestLimitsCountRefusedAtTheirEndpoint(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-release }))
	t.Cleanup(func() {
		srv.CloseClientConnections()
		srv.Close()
	})
	var sinks atomic.Int32
	r := newRig(t, rigConfig{
		url: srv.URL, opts: nethost.Options{MaxOpen: 1, AllowPrivateNetworks: true},
		streams: countingStreams{&sinks},
		conn:    func(c *nethost.Connection) { c.Match = func(_, p string) (string, bool) { return "r" + p, true } },
	})
	r.send("Send", "GET", "/first")
	r.send("Download", "/second", "s")
	if got := r.end(); got != "failed|0|too many open requests||" {
		t.Fatalf("got %q", got)
	}
	if sinks.Load() != 0 {
		t.Error("a refused request must not open its sink")
	}
	obs := r.observed()
	if len(obs) != 1 || obs[0].Route != "r/second" || obs[0].Class != nethost.NotSent {
		t.Errorf("observations: %+v", obs)
	}
	close(release)
	r.end()
}

type countingStreams struct{ sinks *atomic.Int32 }

func (s countingStreams) Source(string) (nethost.Source, error) { return nethost.Source{}, nil }
func (s countingStreams) Sink(string) (nethost.Sink, error) {
	s.sinks.Add(1)
	return &memSink{}, nil
}

// TestHostEndIsNotObserved shows that a request, which the end of the server cuts off, is no network failure.
func TestHostEndIsNotObserved(t *testing.T) {
	entered := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)
	base, _ := url.Parse(srv.URL)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	host := nethost.New(ctx, nethost.Options{AllowPrivateNetworks: true})
	var mu sync.Mutex
	var observed int
	completed := make(chan struct{})
	b := nethost.Binding{
		Resolve: func(string) (nethost.Connection, error) { return nethost.Connection{BaseURL: base}, nil },
		Complete: func(context.Context, gostrict.Detached, []any) error {
			close(completed)
			return nil
		},
		Observe: func(nethost.Observation) {
			mu.Lock()
			defer mu.Unlock()
			observed++
		},
	}
	bundle, err := rigBundle()
	if err != nil {
		t.Fatal(err)
	}
	pool, err := gostrict.NewPoolFunc(t.Context(), engines()["vm"], bundle, 1, func(*gostrict.Slot) gostrict.Funcs { return host.Funcs(b) })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	if _, err := pool.Call(t.Context(), "Send", "GET", "/x"); err != nil {
		t.Fatal(err)
	}
	<-entered
	cancel()
	select {
	case <-completed:
	case <-time.After(20 * time.Second):
		t.Fatal("the request did not end")
	}
	mu.Lock()
	defer mu.Unlock()
	if observed != 0 {
		t.Errorf("observations: %d", observed)
	}
}
