// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package program_test

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/program"
	"go.wdy.de/nago/presentation/ui/threejs"
)

const app = `package app

import (
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/msg"
)

// Ping goes to the backend.
type Ping struct {
	N    int32
	F    float64
	Tags list.List[string]
	M    dict.Map[string, uint32]
}

// Pong comes from the backend.
type Pong struct {
	N int32
	S string
}

type State struct {
	N int32
}

func Init(n int32) *State { return &State{N: n} }

func OnPong(s *State, p Pong) msg.Out[Ping] {
	s.N += p.N
	out := msg.New[Ping]()
	out.Send(Ping{N: s.N, F: -0.0, Tags: list.Of(p.S), M: dict.New[string, uint32]()})
	return out
}
`

const ping = "llm.local/app.Ping"

func compile(t *testing.T) *gostrict.Bundle {
	b, diags, err := program.Compile("llm.local/app", gostrict.Module{Path: "llm.local/app", FS: fstest.MapFS{"app.go": {Data: []byte(app)}}})
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags)
	}
	return b
}

// TestChannel exchanges messages between the backend and a program, whose frontend the test plays.
func TestChannel(t *testing.T) {
	b := compile(t)
	var received []gostrict.Message

	nt := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("scene", func(wnd core.Window) core.View {
			ch := program.ChannelOf(wnd, "scene")
			return ui.VStack(
				threejs.ThreeJS("scene", b).
					Init("Init", int32(5)).
					OnMessage(func(ctx context.Context, m gostrict.Message) error {
						received = append(received, m)
						return nil
					}),
				ui.PrimaryButton(func() {
					if err := ch.Send(gostrict.Message{Type: "llm.local/app.Pong", Fields: []any{int32(2), "hi"}}); err != nil {
						t.Error(err)
					}
				}).Title("pong"),
				ui.PrimaryButton(func() {
					if err := ch.Send(gostrict.Message{Type: ping, Fields: []any{int32(1)}}); err == nil {
						t.Error("the program does not handle Ping")
					}
				}).Title("wrong"),
			)
		})
	})
	w := nt.Open(t, nil, "scene")

	var listen *proto.CallRequested
	for _, c := range w.AsyncCalls() {
		if l, ok := c.Call.(*proto.GostrictListen); ok && l.Id == "scene" {
			listen = c
		}
	}
	if listen == nil {
		t.Fatal("the component must open its channel")
	}

	// the backend sends a message to the program
	w.Click(w.Find(nagotest.Text("pong")))
	calls := w.AsyncCalls()
	deliver, ok := calls[len(calls)-1].Call.(*proto.GostrictDeliver)
	if !ok || deliver.Id != "scene" || len(deliver.Messages) != 1 || deliver.Messages[0].Type != "llm.local/app.Pong" {
		t.Fatalf("unexpected call %#v", calls[len(calls)-1].Call)
	}
	w.Click(w.Find(nagotest.Text("wrong")))

	// the program answers with a message, whose -0 survives
	resolve := func(msgs ...proto.GostrictMessage) {
		if err := w.Scope().Dispatch(&proto.CallResolved{CallPtr: listen.CallPtr, Ret: &proto.GostrictReceived{Messages: msgs}}); err != nil {
			t.Fatal(err)
		}
		w.Settle()
	}
	resolve(proto.GostrictMessage{Type: ping, Fields: proto.GostrictValues{
		&proto.GostrictInt{V: 7},
		&proto.GostrictNegZero{},
		&proto.GostrictList{V: proto.GostrictValues{&proto.GostrictStr{V: "hi"}}},
		&proto.GostrictList{V: proto.GostrictValues{&proto.GostrictList{V: proto.GostrictValues{&proto.GostrictStr{V: "k"}, &proto.GostrictInt{V: 4000000000}}}}},
	}})
	// a message, which the program does not send, and a broken one are dropped
	resolve(proto.GostrictMessage{Type: "llm.local/app.Pong"})
	resolve(proto.GostrictMessage{Type: ping})

	if len(received) != 1 {
		t.Fatalf("expected one message, got %v", received)
	}
	m := received[0]
	want := []any{int32(7), math.Copysign(0, -1), []any{"hi"}, []any{[]any{"k", uint32(4000000000)}}}
	if m.Type != ping || !gostrict.SameValues(m.Fields, want) {
		t.Fatalf("unexpected message %#v", m)
	}
}

// TestBundleEndpoint serves the manifest of a registered bundle under its immutable URL.
func TestBundleEndpoint(t *testing.T) {
	b := compile(t)
	uri := program.Register(b)
	if !strings.HasPrefix(string(uri), program.BundlePath) || !strings.Contains(string(uri), b.Hash) {
		t.Fatalf("unexpected uri %s", uri)
	}

	rec := httptest.NewRecorder()
	program.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, string(uri), nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("unexpected response %d %v", rec.Code, rec.Header())
	}
	body := rec.Body.String()
	for _, s := range []string{`"hash":"` + b.Hash, `"handlers":{"llm.local/app.Pong":"OnPong"}`, `"kind":"state"`, `"js":"(function`} {
		if !strings.Contains(body, s) {
			t.Fatalf("the manifest lacks %s:\n%s", s, body)
		}
	}

	rec = httptest.NewRecorder()
	program.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, program.BundlePath+"unknown.json", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// TestEngine calls a program through the engine of the browser, whose answers the test gives.
func TestEngine(t *testing.T) {
	b := compile(t)
	type result struct {
		values []any
		err    error
	}
	results := make(chan result, 4)

	nt := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("engine", func(wnd core.Window) core.View {
			return ui.PrimaryButton(func() {
				// the calls wait for the browser, so they run in a goroutine of their own
				go func() {
					in, err := program.Engine{Window: wnd}.Load(context.Background(), b, nil)
					if err != nil {
						results <- result{err: err}
						return
					}
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					res, err := in.Call(ctx, "Init", int32(3))
					results <- result{res, err}
					if err != nil {
						return
					}
					res, err = in.Call(ctx, "OnPong", res[0], []any{int32(4), "x"})
					results <- result{res, err}
				}()
			}).Title("run")
		})
	})
	w := nt.Open(t, nil, "engine")
	w.Click(w.Find(nagotest.Text("run")))

	// answer awaits the next call of the program and answers it like the browser
	answered := map[proto.Ptr]bool{}
	answer := func(ret *proto.GostrictReturn) *proto.GostrictCall {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			for _, c := range w.AsyncCalls() {
				if gc, ok := c.Call.(*proto.GostrictCall); ok && !answered[c.CallPtr] {
					answered[c.CallPtr] = true
					if err := w.Scope().Dispatch(&proto.CallResolved{CallPtr: c.CallPtr, Ret: ret}); err != nil {
						t.Fatal(err)
					}
					return gc
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatal("no call of the program")
		return nil
	}

	call := answer(&proto.GostrictReturn{Values: proto.GostrictValues{&proto.GostrictRef{V: 1}}})
	if call.Name != "Init" || len(call.Args) != 1 {
		t.Fatalf("unexpected call %#v", call)
	}
	r := <-results
	if r.err != nil {
		t.Fatal(r.err)
	}
	if _, ok := r.values[0].(gostrict.Token); !ok {
		t.Fatalf("expected a state, got %#v", r.values)
	}

	call = answer(&proto.GostrictReturn{Values: proto.GostrictValues{&proto.GostrictList{V: proto.GostrictValues{
		&proto.GostrictList{V: proto.GostrictValues{&proto.GostrictInt{V: 7}, &proto.GostrictFloat{V: 1.5}, &proto.GostrictList{}, &proto.GostrictList{}}},
	}}}})
	if ref, ok := call.Args[0].(*proto.GostrictRef); !ok || ref.V != 1 {
		t.Fatalf("the state must be passed as its reference: %#v", call.Args)
	}
	r = <-results
	if r.err != nil {
		t.Fatal(r.err)
	}
	want := []any{[]any{[]any{int32(7), 1.5, []any{}, []any{}}}}
	if !gostrict.SameValues(r.values, want) {
		t.Fatalf("unexpected results %#v", r.values)
	}
}
