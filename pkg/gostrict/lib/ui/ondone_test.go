// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/net/nethost"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/std/task/taskhost"
	"go.wdy.de/nago/presentation/core"
)

// TestOnDone runs the whole way of the design: a view starts a request, the function of the request runs as a call of
// its own in another instance and only after it the view changes its state on the event loop, which renders it again.
func TestOnDone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail" {
			w.WriteHeader(http.StatusTeapot)
			return
		}
		fmt.Fprint(w, "hello")
	}))
	t.Cleanup(srv.Close)
	base, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	b := compileView(t, `package v

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/lib/net"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
)

func button(wnd ui.Window, title, path string, status ui.StringState) ui.TButton {
	return ui.PrimaryButton(func() {
		id := net.Send("api", net.NewTextRequest("GET", path, ""), func(r net.Result) error {
			if r.Outcome != net.OK {
				return errors.New(string(r.Outcome) + " for " + path)
			}
			return nil
		})
		ui.OnDone(wnd, id, func(failure string) {
			status.Set("ended[" + failure + "]")
		})
	}).Title(title)
}

func Render(wnd ui.Window) ui.View {
	status := ui.StringStateOf(wnd, "status")
	return ui.VStack(ui.Text("status:"+status.Get()), button(wnd, "ok", "/ok", status), button(wnd, "fail", "/fail", status))
}
`)

	host := nethost.New(t.Context(), nethost.Options{AllowPrivateNetworks: true})
	resolve := func(name string) (nethost.Connection, error) {
		if name != "api" {
			return nethost.Connection{}, errors.New("unknown")
		}
		return nethost.Connection{BaseURL: base}, nil
	}

	var pool *gostrict.Pool
	funcs := func(wnd core.Window) gostrict.Funcs {
		res := ui.Host()
		maps.Copy(res, host.Funcs(nethost.Binding{Resolve: resolve, Complete: nethost.PoolCompleter(pool), Jobs: func() *taskhost.Jobs {
			if wnd == nil {
				return nil
			}
			return taskhost.Of(wnd)
		}}))
		return res
	}
	pool, err = gostrict.NewPool(t.Context(), &vm.Engine{}, b, funcs(nil), 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })

	configure := func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.ondone")
		cfg.RootView("v", func(wnd core.Window) core.View {
			in, err := (&vm.Engine{}).Load(context.Background(), b, funcs(wnd))
			if err != nil {
				t.Fatal(err)
			}
			return ui.Render(context.Background(), in, "Render", wnd)
		})
	}

	w := nagotest.New(t, configure).Open(t, nil, "v")
	w.Find(nagotest.Text("status:"))

	await := func(text string) {
		t.Helper()
		for range 400 {
			w.Settle()
			if w.FindAll(nagotest.Text(text)).Len() == 1 {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		t.Fatalf("%q was not rendered: %#v", text, w.Tree())
	}

	w.Click(w.Find(nagotest.Text("ok")))
	await("status:ended[]")

	w.Click(w.Find(nagotest.Text("fail")))
	await("status:ended[rejected for /fail]")
}
