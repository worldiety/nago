// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package library holds the functional requirements of the library domain.
package library

import "github.com/worldiety/speclink/spec"

var RLibStock = spec.Declare(spec.Requirement{
	ID:         "R-LIB-STOCK",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Bestand einsehen",
	Text:       "Der Bestand MUSS mit Titel, Autor, Gesamtzahl der Exemplare und den aktuellen Ausleihern eingesehen werden können.",
})
