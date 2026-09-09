// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package dec holds the recorded decisions of this example.
//
// One requirement per file, the file named after the identity. spec.Declare puts it into the runtime
// catalogue, which is what lets the running program - and the assistant in it - explain itself. Without that
// call the value exists only for the static tool, and Go offers no reflection over package level variables
// to find it again.
package dec

import "github.com/worldiety/speclink/spec"

var RDecAvailability = spec.Declare(spec.Requirement{
	ID:         "R-DEC-AVAILABILITY",
	Kind:       spec.Decision,
	Discipline: spec.Technical,
	Status:     spec.Normative,
	Title:      "Verfügbarkeit wird berechnet, nicht gespeichert",
	Text:       "Die Zahl der freien Exemplare ergibt sich aus Copies minus der Länge von LentTo und wird nirgends abgelegt.",
	Rationale:  "Ein abgeleiteter Wert, der zusätzlich gespeichert wird, ist ein Wert, der sich mit sich selbst widersprechen kann. Genau das passiert beim ersten Absturz zwischen zwei Schreibvorgängen, und danach glaubt niemand mehr der Zahl.",
	// Mandatory for a decision, and the half nobody writes unprompted: a justification is pleasant to write
	// and gets written at length, while admitting what the ruling makes worse does not.
	Consequences: "Jede Anzeige rechnet neu. Eine Auswertung über zehntausend Titel lässt sich nicht über einen Index beantworten, sondern muss den Bestand lesen.",
})
