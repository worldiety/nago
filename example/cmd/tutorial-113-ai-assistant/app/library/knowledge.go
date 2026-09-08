// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

import (
	"fmt"
	"strings"
)

// This file makes the reasoning behind the library readable at run time, so an assistant can answer the one
// question no domain tool can: why is it like this.
//
// # Why that is worth a tool of its own
//
// A person standing in front of a refusal they did not expect - "von Der Prozess ist derzeit kein Exemplar
// frei" - has two questions, and only the first is about data. The second is whether the system is right to
// refuse, and answering it needs the reasoning, not the record. An assistant that can only read the stock
// will invent an explanation, confidently and plausibly, which is worse than saying nothing.
//
// # Why this list is the weak version
//
// The decisions below are maintained by hand, and that is exactly the thing that rots. Nothing breaks when
// somebody changes Available() and leaves DecCopiesDerived describing the old behaviour; the assistant then
// explains a rule that no longer exists, with the full authority of having looked it up.
//
// A project that uses speclink (https://github.com/worldiety/speclink) does not have this problem, because
// the reasoning is not a copy:
//
//   - Requirements live in requirements/dec/R-DEC-<NAME>.spec.go as ordinary Go values. A Decision must
//     carry both a Rationale and its Consequences, or the build fails.
//   - spec.For[LendBook](spec.Satisfies(...), spec.Help("...")) sits in uc_lend_book.annotation.go, in the
//     normal build. Delete the use case and that file stops compiling.
//   - Both are readable at run time - the requirement tree with go:embed, the annotations through
//     spec.Entries() - so the assistant reads what the build already verified rather than a second copy.
//
// See the tutorial text for how that is wired. The shape of the tools does not change; only where the
// content comes from does.

// Decision records why the library works the way it does.
//
// The three fields mirror what speclink demands of a Kind: Decision, and the third is the one that matters:
// a justification is pleasant to write and gets written at length, while admitting what the decision makes
// worse does not - and without it the record reads as advocacy rather than as an account. It is also what
// stops somebody in three years from starting an improvement that was already considered and paid for.
type Decision struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Text         string `json:"text" desc:"was entschieden wurde, in einem Satz"`
	Rationale    string `json:"rationale" desc:"warum es so entschieden wurde"`
	Consequences string `json:"consequences" desc:"was die Entscheidung kostet — der Teil, den niemand ungefragt aufschreibt"`
}

// Capability is one thing the library can do, in the words a user needs, together with the decisions it
// rests on.
//
// Under speclink this is not a list at all: it is spec.Help written at the use case itself, and
// spec.Satisfies naming the requirement. Both are read back with spec.Entries().
type Capability struct {
	UseCase   string   `json:"useCase" desc:"Name des Use Case im Code"`
	Tool      string   `json:"tool" desc:"Name des Werkzeugs, über das der Assistent ihn erreicht"`
	Help      string   `json:"help" desc:"was die Fähigkeit tut, für einen Menschen erklärt"`
	Decisions []string `json:"decisions" desc:"Kennungen der Entscheidungen, auf denen sie beruht"`
}

// Decisions returns every recorded decision, in a stable order.
func Decisions() []Decision {
	return []Decision{
		{
			ID:           "DEC-AVAILABILITY-DERIVED",
			Title:        "Verfügbarkeit wird berechnet, nicht gespeichert",
			Text:         "Die Zahl der freien Exemplare ergibt sich aus Copies minus der Länge von LentTo und wird nirgends abgelegt.",
			Rationale:    "Ein abgeleiteter Wert, der zusätzlich gespeichert wird, ist ein Wert, der sich mit sich selbst widersprechen kann. Genau das passiert beim ersten Absturz zwischen zwei Schreibvorgängen, und danach glaubt niemand mehr der Zahl.",
			Consequences: "Jede Anzeige rechnet neu. Eine Auswertung über zehntausend Titel lässt sich nicht über einen Index beantworten, sondern muss den Bestand lesen.",
		},
		{
			ID:           "DEC-BORROWER-IS-A-NAME",
			Title:        "Ausleiher werden als Name geführt, nicht als Benutzerkonto",
			Text:         "LentTo enthält Klartextnamen; es gibt keine Verknüpfung zur Benutzerverwaltung.",
			Rationale:    "Eine Bibliothek leiht auch an Menschen aus, die kein Konto in diesem System haben — Gäste, Praktikanten, die Kollegin aus dem Nachbarhaus. Ein Fremdschlüssel auf die Benutzertabelle hätte genau diese Fälle ausgeschlossen.",
			Consequences: "Zwei Personen gleichen Namens sind nicht unterscheidbar, und es gibt keine automatische Erinnerung, weil das System keine Adresse kennt. Wer das braucht, muss diese Entscheidung zuerst umdrehen.",
		},
		{
			ID:           "DEC-RETURN-MATCHES-LOOSELY",
			Title:        "Die Rücknahme vergleicht Namen ohne Rücksicht auf Groß- und Kleinschreibung",
			Text:         "ReturnBook findet den Ausleiher über einen Vergleich, der Groß- und Kleinschreibung ignoriert.",
			Rationale:    "Der Name wird von einem Menschen getippt — und zunehmend von einem Sprachmodell aus einem Satz gebildet. Ein exakter Vergleich hätte zur Folge, dass „anna“ ein Exemplar nicht zurückgeben kann, das „Anna“ ausgeliehen hat, und der Fehler sähe aus wie ein Datenfehler.",
			Consequences: "„Anna“ und „anna“ können nicht gleichzeitig je ein Exemplar desselben Titels halten, ohne dass die Rücknahme rät. Bei mehr als einer Handvoll Ausleihern braucht es eine echte Kennung statt eines Namens.",
		},
	}
}

// Capabilities returns what the library can do and what each of those rests on.
func Capabilities() []Capability {
	return []Capability{
		{
			UseCase:   "FindAllBooks",
			Tool:      "list_books",
			Help:      "Zeigt den Bestand mit Titel, Autor, Gesamtzahl der Exemplare und den aktuellen Ausleihern. Die Zahl der freien Exemplare ergibt sich aus den ersten beiden.",
			Decisions: []string{"DEC-AVAILABILITY-DERIVED"},
		},
		{
			UseCase:   "LendBook",
			Tool:      "lend_book",
			Help:      "Gibt ein Exemplar an eine Person heraus. Ist keines frei, wird die Ausleihe abgelehnt — es gibt keine Vormerkung.",
			Decisions: []string{"DEC-AVAILABILITY-DERIVED", "DEC-BORROWER-IS-A-NAME"},
		},
		{
			UseCase:   "ReturnBook",
			Tool:      "return_book",
			Help:      "Nimmt ein Exemplar von einer Person zurück. Hat die Person kein Exemplar dieses Titels, wird die Rücknahme abgelehnt.",
			Decisions: []string{"DEC-BORROWER-IS-A-NAME", "DEC-RETURN-MATCHES-LOOSELY"},
		},
	}
}

// FindDecision returns the decision with the given identity.
func FindDecision(id string) (Decision, bool) {
	for _, d := range Decisions() {
		if strings.EqualFold(d.ID, id) {
			return d, true
		}
	}

	return Decision{}, false
}

// DecisionIndex renders every decision as one line.
//
// The index goes into the system prompt and the full text stays behind a tool. That split is the whole trick:
// the model needs to know what exists in order to decide what is worth looking up, and it does not need the
// rationale of every decision on every single turn. A prompt carrying all of them grows with the project
// until it crowds out the conversation.
func DecisionIndex() string {
	var sb strings.Builder

	for _, d := range Decisions() {
		fmt.Fprintf(&sb, "%s — %s: %s\n", d.ID, d.Title, d.Text)
	}

	return sb.String()
}
