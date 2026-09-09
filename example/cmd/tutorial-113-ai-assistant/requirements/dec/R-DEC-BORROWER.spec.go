// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dec

import "github.com/worldiety/speclink/spec"

var RDecBorrower = spec.Declare(spec.Requirement{
	ID:           "R-DEC-BORROWER",
	Kind:         spec.Decision,
	Discipline:   spec.Business,
	Status:       spec.Normative,
	Title:        "Ausleiher werden als Name geführt, nicht als Benutzerkonto",
	Text:         "LentTo enthält Klartextnamen; es gibt keine Verknüpfung zur Benutzerverwaltung.",
	Rationale:    "Eine Bibliothek leiht auch an Menschen aus, die kein Konto in diesem System haben — Gäste, Praktikanten, die Kollegin aus dem Nachbarhaus. Ein Fremdschlüssel auf die Benutzertabelle hätte genau diese Fälle ausgeschlossen.",
	Consequences: "Zwei Personen gleichen Namens sind nicht unterscheidbar, und es gibt keine automatische Erinnerung, weil das System keine Adresse kennt. Wer das braucht, muss diese Entscheidung zuerst umdrehen.",
})
