// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package ui creates the views of nago in a program, which runs on a server, e.g. in goja. It mirrors the package
// presentation/ui of nago: the same constructors, builder methods, types and constants, with int32 instead of int.
// Only the server implements its functions, so such a program does not run in a browser.
//
// A program renders a window by an exported function, which takes the window and returns the view:
//
//	func Render(wnd ui.Window) ui.View {
//		name := ui.StringStateOf(wnd, "name")
//		return ui.VStack(
//			ui.Text("Kunde").Font(ui.HeadlineSmall),
//			ui.TextField("Name", name.Get()).InputValue(name),
//			ui.PrimaryButton(func() { save(name.Get()) }).Title("Speichern"),
//		).Gap(ui.L16).Padding(ui.Padding{}.All(ui.L16))
//	}
//
// Each builder method returns its own type, also where nago returns a DecoredView, so that the calls chain. A view
// may take functions of the program, e.g. the action of a button or the function, which renders the cells of a
// sheet. The window calls them later, when the user clicks or the window renders, and then renders the window
// again. A state of the window, e.g. StringStateOf, keeps a value between two renders.
//
//gostrict:header
package ui

import "go.wdy.de/nago/pkg/gostrict/std/js"

// View is any view, e.g. a TText or a TStack, which a container takes.
type View interface{ isView() }

// Window is the window of nago, which renders the view.
type Window struct{ ref js.Ref }

// SVG is the source of an SVG image, e.g. an icon.
type SVG string

// URI is a URI, e.g. of a link.
type URI string
