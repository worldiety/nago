// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package pass calls target and passes its results on to the host, also beside an error, whose results a script may
// not read (wokoda R-SCRIPT-GS-OPTIONAL).
package pass

import "go.wdy.de/nago/pkg/gostrict/internal/whole/n10/target"

// Main returns the results of target.
func Main(id string, reject bool) (target.Info, error) {
	return target.Main(id, reject)
}
