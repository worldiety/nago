// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

import "github.com/worldiety/speclink/spec"

var RLibReturn = spec.Declare(spec.Requirement{
	ID:         "R-LIB-RETURN",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Exemplar zurücknehmen",
	Text:       "Ein ausgeliehenes Exemplar MUSS von der Person zurückgenommen werden können, die es hält.",
})
