// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package ailibrary exposes the library to a language model.
//
// It sits beside ui/ rather than inside the context on purpose: a model is a way of reaching this context,
// exactly as a screen is, and neither belongs to the domain. The dependency points inwards only - ailibrary
// imports library, and library imports neither.
//
// That placement is also the answer to the question this example exists for. Because the tools call use
// cases, and a use case audits the acting subject, the assistant is bounded by the permissions of whoever is
// operating it. There is no second authorisation model to write, and none to forget.
package ailibrary

import (
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/app/library"
)

// Tools exposes the library to the model.
//
// Note what is not here: no wrapper around the use cases, no subject parameter, no per-request rebuild. A
// nago use case has the shape func(auth.Subject, Request) (Response, error), which is exactly the shape a
// tool needs, so [completion.NewUseCaseTool] takes it unchanged and passes the acting subject through on
// every call. The tools are therefore built once, at start-up, and shared by every window and every user.
//
// The rules this file follows are worth stating explicitly, because they are what keeps an assistant safe:
//
//  1. A tool goes through a use case, never through a repository. The use case audits the acting subject, so
//     the assistant can only ever do what the person operating it could do by hand. This is not a
//     convention - it is the entire authorization model.
//  2. A tool that changes something is marked with AsMutating. Not described as writing in prose: marked. A
//     sentence in the system prompt is a request the model may ignore; the mark is a gate it cannot pass.
//  3. A listing goes through NewSeqTool, which bounds the result and tells the model when it was cut short.
//     Handing a model an unbounded list is how a context window gets exhausted.
//  4. Names are lower_snake_case and descriptions are written for someone who has never seen the
//     application. Both are validated at construction, so a mistake is a panic at start-up rather than a
//     confused model at runtime.
//  5. The domain types are handed to the model directly, carrying their own `desc` tags. A DTO is written
//     when the model needs something the aggregate does not have - not to rename fields.
//  6. WithResultDoc is used where the shape of the answer is not obvious from its field names. It costs
//     tokens on every request, so it is spent deliberately: the listing wraps its entries in a truncation
//     flag the model has to react to, while lend_book returns a sentence and a number that speak for
//     themselves.
func Tools(uc library.UseCases) []completion.Tool {
	list := completion.NewSeqTool("list_books",
		"Listet den Bestand der Bibliothek mit Titel, Autor, Gesamtzahl der Exemplare und den aktuellen Ausleihern. "+
			"Nutze die Filter, um die Antwort klein zu halten, statt die ganze Liste zu holen und selbst zu suchen.",
		uc.FindAllBooks).
		WithResultDoc()

	lend := completion.NewUseCaseTool("lend_book",
		"Leiht ein Exemplar eines Buches an eine Person aus. Nenne vorher Titel und Person, damit klar ist, was passiert.",
		uc.LendBook).
		AsMutating("gibt ein Exemplar heraus und trägt die Person als Ausleiher ein")

	ret := completion.NewUseCaseTool("return_book",
		"Nimmt ein ausgeliehenes Exemplar von einer Person zurück.",
		uc.ReturnBook).
		AsMutating("nimmt ein Exemplar zurück und entfernt die Person aus der Ausleihliste")

	return []completion.Tool{list, lend, ret}
}

// SystemPrompt is the domain half of what the model is told.
//
// It carries no requirement index and no list of decisions. Both come from the framework: aispeclink.Index
// renders the catalogue, and uicompletion.WindowContext renders where the user stands. Writing either of
// them here would be a copy of something already recorded, and the copy is the one that rots.
//
// What is deliberately absent is any instruction about asking before writing. That is enforced structurally
// by ConfirmMutations, and repeating it here would suggest it were the prompt's job.
const SystemPrompt = `Du hilfst beim Betrieb einer kleinen Bibliothek.

Arbeitsweise:
- Rate nicht. Alles Fachliche steht hinter Werkzeugen; wenn du etwas nicht weißt, hole es dir.
- Antworte knapp und in ganzen Sätzen. Nenne Bücher mit Titel und Autor, nicht mit ihrer Kennung.
- Wenn ein Werkzeug meldet, dass die Liste gekürzt wurde, sage das dazu und grenze die Suche ein, statt so zu tun, als wäre sie vollständig.
- Ein Berechtigungsfehler ist ein erwartetes Ergebnis, keine Störung. Sage dann, dass dem Nutzer die Berechtigung fehlt.

Fragt jemand, WARUM sich das System so verhält oder warum es etwas ablehnt, lies die zugehörige Anforderung
mit read_requirement und antworte daraus. Erfinde keine Begründung — eine erfundene Regel klingt wie das
System, das über sich selbst spricht, und ist schlimmer als keine Antwort. Bei Entscheidungen nenne auch,
was sie kostet, wenn jemand sie in Frage stellt.

Achte auf den Zustand einer Anforderung: „planned" heißt, dass etwas bewusst noch nicht umgesetzt ist. Das
ist etwas anderes als „gibt es nicht", und der Unterschied ist für den Fragenden der ganze Punkt.

Für Orientierungsfragen („was kann das hier eigentlich") nimm read_capabilities.

Bezieht sich jemand auf etwas, das er gerade sieht, oder hast du etwas geändert, sieh mit inspect_screen nach,
was tatsächlich angezeigt wird, statt es anzunehmen.`
