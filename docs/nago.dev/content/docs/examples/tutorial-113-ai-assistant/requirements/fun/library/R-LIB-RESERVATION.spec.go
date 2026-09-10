// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

import "github.com/worldiety/speclink/spec"

// RLibReservation is deliberately Planned and bound to nothing.
//
// It is in this example on purpose. speclink demands a binding only for Normative requirements, so a planned
// one appears in no binding and would be missing from a view built on spec.Entries() alone — and an
// assistant reading that view would answer "so etwas gibt es nicht", which is a different and much worse
// statement than "das ist noch nicht umgesetzt". Reading the catalogue keeps the two apart.
var RLibReservation = spec.Declare(spec.Requirement{
	ID:         "R-LIB-RESERVATION",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Planned,
	Title:      "Vormerkung auf ein ausgeliehenes Exemplar",
	Text:       "Ist kein Exemplar frei, SOLL sich eine Person vormerken lassen können und benachrichtigt werden, sobald eines zurückkommt.",
})
