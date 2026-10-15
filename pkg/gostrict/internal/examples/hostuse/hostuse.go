// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package hostuse calls the functions of a header package.
package hostuse

import (
	"go.wdy.de/nago/pkg/gostrict/internal/examples/host"
)

func Run(a, b float64) (string, float64) {
	r, err := host.Ratio(a, b)
	if err != nil {
		return host.Label(-1) + " " + err.Error(), 0
	}
	return host.Label(7), host.Clamp(host.Mix(a, b)+r, 0, 100)
}
