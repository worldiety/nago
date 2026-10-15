// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package form is a program, which renders a form with many views of lib/ui, for the test of the package.
package form

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

var hint = ui.Border{}.Radius(ui.L4).Width(ui.L1).Color(ui.ColorLine)

// Render renders the window.
func Render(wnd ui.Window) ui.View {
	name := ui.StringStateOf(wnd, "name")
	age := ui.IntStateOf(wnd, "age")
	agree := ui.BoolStateOf(wnd, "agree")
	dark := ui.BoolStateOf(wnd, "dark")
	size := ui.AutoRadioStateGroup(wnd, "size", 3)

	greeting := "Hallo " + name.Get() + " (" + text.FromInt(age.Get()) + ")"
	if agree.Get() {
		greeting += ", einverstanden"
	}

	return ui.ScrollView(
		ui.VStack(
			ui.WindowTitle("Formular"),
			ui.H2("Kunde"),
			ui.TextField("Name", name.Get()).InputValue(name).SupportingText("Vor- und Nachname"),
			ui.IntField("Alter", age.Get(), age),
			ui.PasswordField("Kennwort", "").Frame(ui.Frame{}.FullWidth()),
			ui.CheckboxField("Einverstanden", agree.Get()).InputValue(agree),
			ui.ToggleField("Dunkel", dark.Get()).InputValue(dark),
			ui.RadioButtonField("Klein", size, 0),
			ui.RadioButtonField("Groß", size, 1),
			ui.HLine(),
			ui.Text(greeting).Border(hint).Padding(ui.Padding{}.All(ui.L8)),
			ui.HStack(
				ui.FilledButton(ui.A0, func() { name.Set("") }).Title("Leeren"),
				ui.Spacer(),
				ui.SecondaryButton(func() { age.Set(age.Get() + 1) }).Title("Älter"),
			).Frame(ui.Frame{}.FullWidth()),
			ui.TextLayout(ui.Text("Fließ"), ui.Text("text")),
			ui.BoxAlign(ui.Center, ui.Text("mittig")).Frame(ui.Frame{Height: ui.L48}),
			ui.Table(ui.TableColumn(ui.Text("Feld")), ui.TableColumn(ui.Text("Wert"))).Rows(
				ui.TableRow(ui.TableCell(ui.Text("Name")), ui.TableCell(ui.Text(name.Get()))),
				ui.TableRow(ui.TableCell(ui.Text("Alter")), ui.TableCell(ui.Text(text.FromInt(age.Get())))),
			),
		).Gap(ui.L8).Alignment(ui.Leading),
	).Frame(ui.Frame{}.FullWidth())
}
