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
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/net/nethost"
)

const token = "s3cr3t-t0ken"

// program sends one request and reports its result to the host function of the test.
const program = `package p

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/lib/net"
	"go.wdy.de/nago/pkg/gostrict/std/task"
)

func Get(path string) task.ID {
	return net.Send("crm", net.NewRequest("GET", path, ""), func(r net.Result) error {
		if r.Outcome != net.OK {
			return errors.New(string(r.Outcome) + "|" + path + "|" + r.Reason)
		}
		return errors.New("ok|" + path + "|" + r.Body)
	})
}

func Unknown() task.ID {
	return net.Send("other", net.NewRequest("GET", "/", ""), report)
}

func report(r net.Result) error {
	return errors.New(string(r.Outcome) + "||" + r.Reason)
}

func WithHeader(name, value string) task.ID {
	r := net.NewRequest("GET", "/x", "")
	r.Header.Set(name, value)
	return net.Send("crm", r, report)
}

func Post(body string, timeout int32) task.ID {
	req := net.NewRequest("POST", "/p", body)
	req.Query.Set("a", "b c")
	req.Timeout = timeout
	return net.Send("crm", req, func(r net.Result) error {
		return errors.New(string(r.Outcome) + "|" + r.Body)
	})
}
`

func compile(t *testing.T) *gostrict.Bundle {
	t.Helper()
	prog, err := gostrict.LoadModules("llm.local/p", gostrict.Module{Path: "llm.local/p", FS: fstest.MapFS{"p.go": {Data: []byte(program)}}})
	if err != nil {
		t.Fatal(err)
	}
	if diags := gostrict.Check(prog); len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	b, diags := gostrict.Emit(prog)
	if len(diags) > 0 {
		t.Fatal(diags.Text())
	}
	return b
}

type fixture struct {
	pool *gostrict.Pool
	ends chan string
	host *nethost.Host
}

// setup returns a pool of an engine, whose requests go to the server and whose ends reach f.ends.
func setup(t *testing.T, e gostrict.Engine, opts nethost.Options, srv *httptest.Server) *fixture {
	t.Helper()
	return setupGuarded(t, e, opts, srv, true)
}

func setupGuarded(t *testing.T, e gostrict.Engine, opts nethost.Options, srv *httptest.Server, private bool) *fixture {
	t.Helper()
	opts.AllowPrivateNetworks = private
	base, err := url.Parse(srv.URL + "/api/")
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{ends: make(chan string, 1024)}
	f.host = nethost.New(t.Context(), opts)
	resolve := func(name string) (nethost.Connection, error) {
		if name != "crm" {
			return nethost.Connection{}, errors.New("unknown " + name)
		}
		return nethost.Connection{BaseURL: base, Secrets: []string{token}, Decorate: func(next http.RoundTripper) http.RoundTripper {
			return roundTripper(func(r *http.Request) (*http.Response, error) {
				r = r.Clone(r.Context())
				r.Header.Set("Authorization", "Bearer "+token)
				return next.RoundTrip(r)
			})
		}}, nil
	}
	complete := func(ctx context.Context, done gostrict.Detached, args []any) error {
		res, err := f.pool.CallDetached(ctx, done, args...)
		if err != nil {
			f.ends <- "engine: " + err.Error()
			return err
		}
		err = gostrict.ErrorOf(res[0])
		f.ends <- err.Error()
		return err
	}
	f.pool, err = gostrict.NewPoolFunc(t.Context(), e, compile(t), 4, func(*gostrict.Slot) gostrict.Funcs {
		return f.host.Funcs(nethost.Binding{Resolve: resolve, Complete: complete})
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.pool.Close() })
	return f
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func (f *fixture) call(t *testing.T, name string, args ...any) {
	t.Helper()
	if _, err := f.pool.Call(t.Context(), name, args...); err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) end(t *testing.T) string {
	t.Helper()
	select {
	case s := <-f.ends:
		return s
	case <-time.After(10 * time.Second):
		t.Fatal("the request did not end")
		return ""
	}
}

func engines() map[string]gostrict.Engine {
	return map[string]gostrict.Engine{"vm": &vm.Engine{}, "goja": &gojaengine.Engine{}}
}

func TestSend(t *testing.T) {
	var auth, gotPath, gotQuery, gotBody, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, gotPath, gotQuery, gotMethod = r.Header.Get("Authorization"), r.URL.Path, r.URL.RawQuery, r.Method
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		switch r.URL.Path {
		case "/api/mirror":
			fmt.Fprintf(w, "you sent %s", auth)
		case "/api/nf":
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "no such thing")
		case "/api/mirror-error":
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "bad token %s", token)
		case "/api/away":
			http.Redirect(w, r, "https://elsewhere.example/steal", http.StatusFound)
		case "/api/near":
			http.Redirect(w, r, "/api/mirror", http.StatusFound)
		default:
			fmt.Fprint(w, "hello")
		}
	}))
	t.Cleanup(srv.Close)

	for name, e := range engines() {
		t.Run(name, func(t *testing.T) {
			f := setup(t, e, nethost.Options{}, srv)

			check := func(fn string, want string, args ...any) {
				t.Helper()
				f.call(t, fn, args...)
				if got := f.end(t); got != want {
					t.Errorf("%s%v: got %q, want %q", fn, args, got, want)
				}
			}

			check("Get", "ok|/a/b|hello", "/a/b")
			if gotPath != "/api/a/b" || gotMethod != "GET" || auth != "Bearer "+token {
				t.Errorf("unexpected request: %s %s %q", gotMethod, gotPath, auth)
			}

			check("Get", "ok|/mirror|you sent Bearer ***", "/mirror")
			check("Get", "ok|/near|you sent Bearer ***", "/near")
			check("Get", "rejected|/nf|HTTP 404: no such thing", "/nf")
			check("Get", "rejected|/mirror-error|HTTP 500: bad token ***", "/mirror-error")
			f.call(t, "Get", "/away")
			if got := f.end(t); !strings.HasPrefix(got, "rejected|/away|HTTP 302: ") {
				t.Errorf("a redirect to another origin must not be followed: %q", got)
			}

			check("Post", "ok|hello", "payload", int32(0))
			if gotMethod != "POST" || gotBody != "payload" || gotPath != "/api/p" || gotQuery != "a=b+c" {
				t.Errorf("unexpected request: %s %s %s %q", gotMethod, gotPath, gotQuery, gotBody)
			}

			for _, bad := range []string{"x", "/a/../b", "/a//b", "/a/./b", `/a\b`, "/a\nb"} {
				f.call(t, "Get", bad)
				if got := f.end(t); !strings.HasPrefix(got, "failed|") {
					t.Errorf("the path %q was not refused: %s", bad, got)
				}
			}

			// the text of a path is never an address, so it can not reach another host or a query
			check("Get", "ok|/user@elsewhere.example?a=1#b|hello", "/user@elsewhere.example?a=1#b")
			if gotPath != "/api/user@elsewhere.example?a=1#b" || gotQuery != "" {
				t.Errorf("the path was interpreted: %q %q", gotPath, gotQuery)
			}

			for _, h := range []string{"authorization", "Cookie", "Host", "Proxy-Foo", "x-forwarded-for"} {
				f.call(t, "WithHeader", h, "x")
				if got := f.end(t); !strings.HasPrefix(got, "failed||the header") {
					t.Errorf("the header %q was not refused: %s", h, got)
				}
			}

			f.call(t, "Unknown")
			if got := f.end(t); got != "failed||the connection is not available" {
				t.Errorf("unexpected end of an unknown connection: %s", got)
			}
		})
	}
}

func TestOutcomeOfTimeoutAndLostConnection(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	t.Cleanup(func() {
		close(release)
		srv.Close()
	})
	f := setup(t, &vm.Engine{}, nethost.Options{}, srv)
	f.call(t, "Post", "x", int32(50))
	if got := f.end(t); got != "unknown|" {
		t.Errorf("a timeout after sending must be unknown: %q", got)
	}

	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	f = setup(t, &vm.Engine{}, nethost.Options{}, closed)
	f.call(t, "Post", "x", int32(0))
	if got := f.end(t); got != "failed|" {
		t.Errorf("an unreachable server must be failed: %q", got)
	}
}

func TestLimits(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	t.Cleanup(func() {
		srv.CloseClientConnections()
		srv.Close()
	})
	f := setup(t, &vm.Engine{}, nethost.Options{MaxOpen: 2}, srv)
	f.call(t, "Post", "a", int32(0))
	f.call(t, "Post", "b", int32(0))
	f.call(t, "Get", "/c")
	if got := f.end(t); got != "failed|/c|too many open requests" {
		t.Errorf("the third request must be refused: %q", got)
	}
	close(release)
	f.end(t)
	f.end(t)
}

func TestMaxBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, strings.Repeat("x", 100))
	}))
	t.Cleanup(srv.Close)
	f := setup(t, &vm.Engine{}, nethost.Options{MaxBody: 10}, srv)
	f.call(t, "Get", "/big")
	if got := f.end(t); got != "unknown|/big|the answer is too large" {
		t.Errorf("unexpected end: %q", got)
	}
}

// TestRequestsRunConcurrently sends requests from all instances, while the ends run in the same instances. Run it
// with -race.
func TestRequestsRunConcurrently(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.Path)
	}))
	t.Cleanup(srv.Close)
	f := setup(t, &vm.Engine{}, nethost.Options{MaxOpen: 4096}, srv)

	const n = 400
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() { f.call(t, "Get", fmt.Sprintf("/%d", i)) })
	}
	seen := map[string]bool{}
	for range n {
		seen[f.end(t)] = true
	}
	wg.Wait()
	if len(seen) != n {
		t.Fatalf("expected %d different ends, got %d", n, len(seen))
	}
}

// TestGuardChecksTheResolvedAddress connects to the server on the local machine, which only the option allows.
func TestGuardChecksTheResolvedAddress(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1) }))
	t.Cleanup(srv.Close)
	f := setupGuarded(t, &vm.Engine{}, nethost.Options{}, srv, false)
	f.call(t, "Get", "/x")
	if got := f.end(t); got != "failed|/x|the server can not be reached" {
		t.Errorf("unexpected end: %q", got)
	}
	if hits.Load() != 0 {
		t.Error("the request reached the forbidden address")
	}
}

func TestPanicOfADecoratorEndsTheRequest(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	base, _ := url.Parse(srv.URL)
	h := nethost.New(t.Context(), nethost.Options{AllowPrivateNetworks: true})
	ends := make(chan string, 1)
	var pool *gostrict.Pool
	var err error
	pool, err = gostrict.NewPool(t.Context(), &vm.Engine{}, compile(t), h.Funcs(nethost.Binding{
		Resolve: func(string) (nethost.Connection, error) {
			return nethost.Connection{BaseURL: base, Decorate: func(http.RoundTripper) http.RoundTripper {
				return roundTripper(func(*http.Request) (*http.Response, error) { panic("defect") })
			}}, nil
		},
		Complete: func(ctx context.Context, done gostrict.Detached, args []any) error {
			res, err := pool.CallDetached(ctx, done, args...)
			if err != nil {
				return err
			}
			ends <- gostrict.ErrorOf(res[0]).Error()
			return nil
		},
	}), 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	if _, err := pool.Call(t.Context(), "Get", "/x"); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-ends:
		if got != "unknown|/x|the request has failed" {
			t.Errorf("unexpected end: %q", got)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the request did not end")
	}
}

func TestForbiddenAddress(t *testing.T) {
	for _, tt := range []struct {
		ip      string
		private bool
		want    bool
	}{
		{"8.8.8.8", false, false},
		{"127.0.0.1", false, true},
		{"127.0.0.1", true, false},
		{"10.0.0.1", false, true},
		{"192.168.1.1", true, false},
		{"100.64.0.1", false, true},
		{"169.254.169.254", true, true},
		{"100.100.100.200", true, true},
		{"168.63.129.16", true, true},
		{"fd00:ec2::254", true, true},
		{"fe80::1", true, true},
		{"::ffff:169.254.169.254", true, true},
		{"0.0.0.0", true, true},
	} {
		if got := nethost.ForbiddenAddress(netip.MustParseAddr(tt.ip), tt.private); got != tt.want {
			t.Errorf("%s private=%v: got %v", tt.ip, tt.private, got)
		}
	}
}
