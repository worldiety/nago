// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package c2 is a link of a chain of use cases, which exceeds the depth of the frames.
package c2

import "go.wdy.de/nago/pkg/gostrict/internal/whole/chain/c3"

// Main calls the next link.
func Main(n int32) int32 {
	return c3.Main(n+1) + 1
}
