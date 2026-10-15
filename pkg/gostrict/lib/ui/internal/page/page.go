// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package page is a program, which renders a window with views of lib/ui, for the test of the package.
package page

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

var card = ui.Border{}.Radius(ui.L8).Width(ui.L1).Color(ui.M8)

// Render renders the window.
func Render(wnd ui.Window) ui.View {
	open := ui.BoolStateOf(wnd, "open")
	title := ui.Text("zu")
	if open.Get() {
		title = ui.Text("offen")
	}
	return ui.VStack(
		title.Font(ui.HeadlineSmall).Color(ui.A0),
		ui.PrimaryButton(func() { open.Set(!open.Get()) }).Title("Umschalten"),
		ui.HStack(ui.Text("links"), ui.Text("rechts").Padding(ui.Padding{}.Horizontal(ui.L4))).
			Gap(ui.L16).
			Border(card).
			WithFrame(func(f ui.Frame) ui.Frame { return f.FullWidth() }),
		ui.Text("Breite").Frame(ui.Frame{Width: ui.Full}),
	).Gap(ui.L16).Padding(ui.Padding{}.All(ui.L8))
}
