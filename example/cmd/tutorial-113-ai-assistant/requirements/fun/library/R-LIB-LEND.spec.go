// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

import "github.com/worldiety/speclink/spec"

var RLibLend = spec.Declare(spec.Requirement{
	ID:         "R-LIB-LEND",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Exemplar ausleihen",
	Text:       "Ein freies Exemplar MUSS an eine namentlich genannte Person ausgeliehen werden können; ist keines frei, MUSS die Ausleihe abgelehnt werden.",
})
