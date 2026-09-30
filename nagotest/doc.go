// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package nagotest runs a nago application within the test process and drives its windows like a user.
//
// Windows opened by [App.Open] involve no HTTP server, no websocket and no protocol decoding: each [Window] is
// connected directly to the scope of the application and receives the rendered trees as Go values.
// Alternatively, [Dial] connects a Window through a websocket to a running server, e.g. one started by
// [Serve], using the identical API. The load package builds upon that to put a server under load. The window emulates the
// frontend, thus it follows navigation requests and records side effects like downloads, file import
// requests, focus requests and asynchronous frontend calls.
//
// A typical table test looks like this:
//
//	app := nagotest.New(t, func(cfg *application.Configurator) {
//		cfg.SetApplicationID("de.worldiety.demo")
//		cfg.RootView("demo", myPage)
//	})
//
//	w := app.Open(t, subject, "demo", nagotest.Values(core.Values{"id": "42"}))
//	w.Type(w.Find(nagotest.Label("Name")), "Torben")
//	w.Click(w.Find(nagotest.Text("Speichern")))
//	w.Find(nagotest.Text("gespeichert")).Visible()
//	w.FindAll(nagotest.Type[*proto.Modal]()).None()
//
// All actions settle the window afterward, which means that they wait until the event loop is idle, all
// pending state changes have been rendered, background tasks like [core.OnAppear] have completed and
// requested navigations have been followed. There is no fake clock, thus delayed functions
// (see [core.Window.PostDelayed]) are not awaited.
//
// Beyond clicking and typing, a window can complete file imports ([Window.Upload]), deliver input events
// to listeners ([Window.Input]), inspect or remount canvas elements ([Window.Canvas]) and act on flow charts
// ([Window.ClickFlowChartNode], [Window.FlowChartAction]).
//
// An action refers to the callbacks of the latest tree. If the application renders again in between, the
// backend discards the action as stale, so the window finds the selection again and repeats it, see
// [StaleRetries]. A view which changes a state during each render renders endlessly, thus the window does not
// settle and the test fails.
//
// Dialogs and notifications are part of the rendered tree, e.g. [proto.Modal] or [proto.AlertNotifications],
// and are found like any other component.
package nagotest
