// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// This file sits next to uc_find_all_books.go and is part of the normal build. That coupling is the point:
// delete the use case and this file stops compiling, so the binding cannot rot unnoticed.
//
// Satisfies is the one statement no static analysis can infer — which requirement a construct was written
// for. Help is what an assistant reads back through read_capabilities.
package library

import (
	"github.com/worldiety/speclink/spec"
	"go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/requirements/dec"
	fun "go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/requirements/fun/library"
)

var _ = spec.For[FindAllBooks](
	spec.Satisfies(fun.RLibStock, dec.RDecAvailability),
	spec.Help("Zeigt den Bestand mit Titel, Autor, Gesamtzahl der Exemplare und den aktuellen Ausleihern. Die Zahl der freien Exemplare ergibt sich aus den ersten beiden."),
)
