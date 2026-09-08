// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package uilibrary renders the library for a human.
//
// The directory is always ui/, so the package name is what tells a reader whose interface this is. It imports
// the context; the context must never import it, because that would invert the ordering and make the domain
// untestable without a renderer.
package uilibrary

import (
	"strconv"

	"go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/app/library"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// Pages are the routes of this context.
type Pages struct {
	Books core.NavigationPath
}

// PageBooks lists the stock.
//
// It is here so the assistant has a screen to be aware of, and so the application.Purpose registered in cfg/
// describes something real. It is deliberately plain: the subject of this example is next door in ai/.
func PageBooks(wnd core.Window, uc library.UseCases) core.View {
	var rows []core.View
	for book, err := range uc.FindAllBooks(wnd.Subject(), library.BookFilter{}) {
		if err != nil {
			return ui.Text(err.Error())
		}

		rows = append(rows, ui.HStack(
			ui.Text(book.Title).Font(ui.BodyMedium),
			ui.Text(book.Author),
			ui.Text(availabilityOf(book)),
		).Gap(ui.L16).FullWidth().Alignment(ui.Leading))
	}

	if len(rows) == 0 {
		rows = append(rows, ui.Text("Kein Bestand sichtbar. Fehlt die Rolle „Bibliothekar“?"))
	}

	return ui.VStack(
		ui.H1("Bestand"),
		ui.Text("Frag den Assistenten unten rechts, was verfügbar ist, oder lass ihn ein Exemplar ausleihen. Vor jeder Änderung wird nachgefragt."),
		ui.VStack(rows...).Gap(ui.L8).FullWidth(),
	).Gap(ui.L16).Alignment(ui.Leading).FullWidth().
		Padding(ui.Padding{}.All(ui.L16))
}

// availabilityOf says in words what a bare number would leave the reader to work out.
func availabilityOf(b library.Book) string {
	switch {
	case b.Available() <= 0:
		return "alle Exemplare ausgeliehen"
	case b.Available() == 1:
		return "1 Exemplar frei"
	default:
		return strconv.Itoa(b.Available()) + " Exemplare frei"
	}
}
