// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"strings"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// A browser authenticates its session by the http-only cookie of the websocket handshake. A client which sends
// its own SessionAssigned must not switch the scope to another session, bypassing the cookie.
func TestClientCannotChooseSessionOverTheWire(t *testing.T) {
	base := Serve(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.wiresessiontest")
		page := func(wnd core.Window) core.View {
			return ui.Text("session: " + string(wnd.Session().ID()))
		}
		cfg.RootView(".", page)
		cfg.RootView("other", page)
	})

	cookie := strings.Repeat("c", 40)
	chosen := strings.Repeat("x", 40)

	w := Dial(t, base, ".", Session(cookie))
	w.Find(Text("session: " + cookie))

	if err := w.tr.send(w, &proto.SessionAssigned{SessionID: proto.Str(chosen)}); err != nil {
		t.Fatal(err)
	}

	// a new window of the scope reads the session again
	w.allocate(Route{Path: "other"})
	w.Settle()
	w.Find(Text("session: " + cookie))
	w.FindAll(Text("session: " + chosen)).None()
}

// A client without a cookie, like a native app or the local transport of this package, assigns its session.
func TestClientWithoutCookieAssignsSession(t *testing.T) {
	app := New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.wiresessiontest")
		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.Text("session: " + string(wnd.Session().ID()))
		})
	})

	sid := strings.Repeat("n", 40)
	app.Open(t, nil, ".", Session(sid)).Find(Text("session: " + sid))
}
