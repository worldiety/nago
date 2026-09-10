// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package library

import (
	"github.com/worldiety/speclink/spec"
	"go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/requirements/dec"
	fun "go.wdy.de/nago/example/cmd/tutorial-113-ai-assistant/requirements/fun/library"
)

var _ = spec.For[LendBook](
	spec.Satisfies(fun.RLibLend, dec.RDecBorrower),
	spec.Help("Gibt ein Exemplar an eine Person heraus. Ist keines frei, wird die Ausleihe abgelehnt — eine Vormerkung gibt es noch nicht."),
)
