// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/canvas"
)

func configure(cfg *application.Configurator) {
	cfg.SetApplicationID("de.worldiety.nagotest")

	cfg.RootView("counter", func(wnd core.Window) core.View {
		count := core.AutoState[int](wnd)
		name := core.AutoState[string](wnd)
		greeting := core.AutoState[string](wnd)
		loaded := core.AutoState[string](wnd)

		core.OnAppear(wnd, "", func(ctx context.Context) {
			loaded.Set("loaded")
		})

		return ui.VStack(
			ui.Text(fmt.Sprintf("count: %d", count.Get())),
			ui.PrimaryButton(func() { count.Set(count.Get() + 1) }).Title("increment"),
			ui.SecondaryButton(nil).Title("disabled").Enabled(false),
			ui.TextField("Name", name.Get()).InputValue(name).KeydownEnter(func() {
				greeting.Set("hello " + name.Get())
			}),
			ui.Text(greeting.Get()),
			ui.Text(loaded.Get()),
			ui.Text(fmt.Sprintf("user: %s valid=%v", wnd.Subject().Name(), wnd.Subject().Valid())),
			ui.PrimaryButton(func() {
				wnd.Navigation().ForwardTo("details", core.Values{"id": "42"})
			}).Title("details"),
			ui.PrimaryButton(func() {
				wnd.ExportFiles(core.ExportFileBytes("a.txt", []byte("hello")))
			}).Title("download"),
			ui.PrimaryButton(func() {
				wnd.RequestFocus("name")
			}).Title("focus"),
		)
	})

	cfg.RootView("details", func(wnd core.Window) core.View {
		return ui.VStack(
			ui.Text("details of "+wnd.Values()["id"]),
			ui.PrimaryButton(func() { wnd.Navigation().Back() }).Title("back"),
		)
	})
}

func TestWindow(t *testing.T) {
	for _, roundTrip := range []bool{false, true} {
		t.Run(fmt.Sprintf("roundTrip=%v", roundTrip), func(t *testing.T) {
			var opts []nagotest.Option
			if roundTrip {
				opts = append(opts, nagotest.WithRoundTrip())
			}

			app := nagotest.New(t, configure, opts...)
			w := app.Open(t, nil, "counter")

			w.Find(nagotest.Text("count: 0")).Visible()
			w.Find(nagotest.Text("loaded"))

			w.Click(w.Find(nagotest.Text("increment")))
			w.Click(w.Find(nagotest.Label("increment")))
			w.Find(nagotest.Text("count: 2"))

			w.Find(nagotest.Label("increment")).Enabled()
			w.FindAll(nagotest.Text("does not exist")).None()
		})
	}
}

func TestInput(t *testing.T) {
	app := nagotest.New(t, configure)
	w := app.Open(t, nil, "counter")

	field := w.Find(nagotest.Label("Name"))
	w.Type(field, "Torben")
	w.PressEnter(w.Find(nagotest.Type[*proto.TextField]()))
	w.Find(nagotest.Text("hello Torben"))
}

func TestNavigation(t *testing.T) {
	app := nagotest.New(t, configure)
	w := app.Open(t, nil, "counter")

	w.Click(w.Find(nagotest.Text("details")))
	w.Find(nagotest.Text("details of 42"))
	if r := w.Route(); r.Path != "details" || r.Values["id"] != "42" {
		t.Fatalf("unexpected route %v", r)
	}

	w.Click(w.Find(nagotest.Text("back")))
	w.Find(nagotest.Text("count: 0"))
	if len(w.History()) != 1 {
		t.Fatalf("unexpected history %v", w.History())
	}
}

func TestSideEffects(t *testing.T) {
	app := nagotest.New(t, configure)
	w := app.Open(t, nil, "counter")

	w.Click(w.Find(nagotest.Text("download")))
	if d := w.Downloads(); len(d) != 1 || d[0].Files[0].Name() != "a.txt" {
		t.Fatalf("unexpected downloads %v", d)
	}

	w.Click(w.Find(nagotest.Text("focus")))
	if w.Focused() != "name" {
		t.Fatalf("unexpected focus %q", w.Focused())
	}
}

func TestSubject(t *testing.T) {
	app := nagotest.New(t, configure)

	w := app.Open(t, nil, "counter")
	anon := w.Find(nagotest.TextContains("user: ")).Text()

	w = app.Open(t, user.SU(), "counter")
	su := w.Find(nagotest.TextContains("user: ")).Text()
	if su == anon || !strings.Contains(su, "valid=true") {
		t.Fatalf("expected su subject but got %q (anon %q)", su, anon)
	}
}

func TestReload(t *testing.T) {
	app := nagotest.New(t, configure)
	w := app.Open(t, nil, "counter")

	w.Click(w.Find(nagotest.Text("increment")))
	w.Reconnect()
	w.Find(nagotest.Text("count: 1"))

	w.Reload()
	w.Find(nagotest.Text("count: 0"))

	w.Close()
}

func TestAsyncCallbackLifecycle(t *testing.T) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("input", func(wnd core.Window) core.View {
			events := core.AutoState[int](wnd)
			answers := core.AutoState[int](wnd)
			closer := core.AutoState[func()](wnd)

			return ui.VStack(
				ui.Text(fmt.Sprintf("events: %d answers: %d", events.Get(), answers.Get())),
				ui.PrimaryButton(func() {
					closer.Set(wnd.AddInputListener("canvas", func(evt core.InputEvent) {
						events.Set(events.Get() + 1)
					}, core.InputEventPointerDown, core.DestroyOnClose))
				}).Title("listen"),
				ui.PrimaryButton(func() { closer.Get()() }).Title("unlisten"),
				ui.PrimaryButton(func() {
					core.AsyncCall(wnd, &proto.CallRequestFocus{ID: "x"}, func(ret proto.CallRet) {
						answers.Set(answers.Get() + 1)
					})
				}).Title("ask"),
			)
		})
	})

	w := app.Open(t, nil, "input")

	// resolve dispatches a frontend answer and reports, if the backend knew the callback
	resolve := func(ptr proto.Ptr, ret proto.CallRet) bool {
		before := len(w.Events())
		if err := w.Scope().Dispatch(&proto.CallResolved{CallPtr: ptr, Ret: ret}); err != nil {
			t.Fatal(err)
		}
		w.Scope().Flush()
		for _, evt := range w.Events()[before:] {
			if _, ok := evt.(*proto.ErrorOccurred); ok {
				return false
			}
		}
		return true
	}

	lastCall := func() *proto.CallRequested {
		calls := w.AsyncCalls()
		return calls[len(calls)-1]
	}

	// a listener resolves many times
	w.Click(w.Find(nagotest.Text("listen")))
	listener := lastCall()
	for range 2 {
		if !resolve(listener.CallPtr, &proto.InputEvent{Type: proto.InputEventType(core.InputEventPointerDown)}) {
			t.Fatal("listener callback must survive a resolution")
		}
	}
	w.Settle()
	w.Find(nagotest.Text("events: 2 answers: 0"))

	// unregistering drops the listener callback and late events are ignored silently
	w.Click(w.Find(nagotest.Text("unlisten")))
	if !resolve(listener.CallPtr, &proto.InputEvent{}) {
		t.Fatal("late input events must not cause an error")
	}
	w.Settle()
	w.Find(nagotest.Text("events: 2 answers: 0"))

	// a conventional call resolves exactly once
	w.Click(w.Find(nagotest.Text("ask")))
	ask := lastCall()
	if !resolve(ask.CallPtr, &proto.RetError{}) {
		t.Fatal("first resolution must find the callback")
	}
	if resolve(ask.CallPtr, &proto.RetError{}) {
		t.Fatal("callback must be removed after the first resolution")
	}
	w.Find(nagotest.Text("events: 2 answers: 1"))
}

func TestUploadInputCanvas(t *testing.T) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest")
		cfg.RootView("io", func(wnd core.Window) core.View {
			uploaded := core.AutoState[string](wnd)
			clicks := core.AutoState[int](wnd)

			ctx := canvas.Context2D(wnd, "board")
			wnd.AddInputListener("board", func(evt core.InputEvent) {
				if evt.Type == core.InputEventPointerDown {
					clicks.Set(clicks.Get() + 1)
				}
				ctx.FillRect(evt.X, evt.Y, 1, 1)
			}, core.InputEventPointerDown, core.InputEventInvalidate)

			return ui.VStack(
				ui.Text("uploaded: "+uploaded.Get()),
				ui.Text(fmt.Sprintf("clicks: %d", clicks.Get())),
				canvas.Canvas("board"),
				ui.PrimaryButton(func() {
					wnd.ImportFiles(core.ImportFilesOptions{ID: "up", OnCompletion: func(files []core.File) {
						var buf strings.Builder
						_, _ = files[0].Transfer(&buf)
						mimeType, _ := files[0].MimeType()
						uploaded.Set(files[0].Name() + " " + mimeType + " " + buf.String())
					}})
				}).Title("import"),
			)
		})
	})

	w := app.Open(t, nil, "io")

	w.Click(w.Find(nagotest.Text("import")))
	if imports := w.Imports(); len(imports) != 1 || imports[0].ID != "up" {
		t.Fatalf("unexpected imports %v", imports)
	}
	w.Upload("up", nagotest.File("a.txt", []byte("hello")))
	w.Find(nagotest.TextContains("uploaded: a.txt text/plain"))
	w.Find(nagotest.TextContains("hello"))

	board := w.Canvas("board")
	board.Commands() // drop everything up to now

	w.Input("board", core.InputEvent{Type: core.InputEventPointerDown, X: 3, Y: 4})
	w.Find(nagotest.Text("clicks: 1"))
	cmds := board.Commands()
	if len(cmds) != 1 {
		t.Fatalf("expected a single command but got %v", cmds)
	}
	if r, ok := cmds[0].(*proto.CanvasFillRect); !ok || r.X != 3 || r.Y != 4 {
		t.Fatalf("unexpected command %#v", cmds[0])
	}

	board.Mount()
	if cmds := board.Commands(); len(cmds) != 1 {
		t.Fatalf("expected a redraw after mount but got %v", cmds)
	}
	w.Find(nagotest.Text("clicks: 1"))
}
