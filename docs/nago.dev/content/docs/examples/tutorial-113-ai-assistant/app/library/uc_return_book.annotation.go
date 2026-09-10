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

var _ = spec.For[ReturnBook](
	spec.Satisfies(fun.RLibReturn, dec.RDecReturnMatch),
	spec.Help("Nimmt ein Exemplar von einer Person zurück. Hat die Person kein Exemplar dieses Titels, wird die Rücknahme abgelehnt."),
)
