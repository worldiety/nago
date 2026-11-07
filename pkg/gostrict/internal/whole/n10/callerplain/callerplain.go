// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package callerplain is an entry point, whose last result is an error, which calls plain, an entry point without an
// error result.
package callerplain

import (
	"go.wdy.de/nago/pkg/gostrict/internal/whole/n10/plain"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Main returns "plain <result of plain>".
func Main(reject bool) (string, error) {
	return "plain " + text.FromInt(plain.Main(reject)), nil
}
