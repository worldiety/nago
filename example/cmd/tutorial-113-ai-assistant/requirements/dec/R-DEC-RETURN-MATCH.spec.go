// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dec

import "github.com/worldiety/speclink/spec"

var RDecReturnMatch = spec.Declare(spec.Requirement{
	ID:           "R-DEC-RETURN-MATCH",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Normative,
	Title:        "Die Rücknahme vergleicht Namen ohne Rücksicht auf Groß- und Kleinschreibung",
	Text:         "ReturnBook findet den Ausleiher über einen Vergleich, der Groß- und Kleinschreibung ignoriert.",
	Rationale:    "Der Name wird von einem Menschen getippt — und zunehmend von einem Sprachmodell aus einem Satz gebildet. Ein exakter Vergleich hätte zur Folge, dass „anna“ ein Exemplar nicht zurückgeben kann, das „Anna“ ausgeliehen hat, und der Fehler sähe aus wie ein Datenfehler.",
	Consequences: "„Anna“ und „anna“ können nicht gleichzeitig je ein Exemplar desselben Titels halten, ohne dass die Rücknahme rät. Bei mehr als einer Handvoll Ausleihern braucht es eine echte Kennung statt eines Namens.",
	DerivedFrom:  []spec.Requirement{RDecBorrower},
})
