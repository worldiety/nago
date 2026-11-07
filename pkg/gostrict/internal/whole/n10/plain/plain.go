// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package plain is an entry point, whose last result is no error.
package plain

import (
	"go.wdy.de/nago/pkg/gostrict/internal/whole/n10/guard"
	"go.wdy.de/nago/pkg/gostrict/internal/whole/store"
)

// Main records its run; with reject, guard.Reject would end it with an error, which it can not return.
func Main(reject bool) int32 {
	store.Put("plain", "ran")
	if reject {
		guard.Reject("rejected plain")
	}
	return 1
}
