// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package wipe is a use case of the area admin, which the user of the test may not call.
package wipe

import "go.wdy.de/nago/pkg/gostrict/internal/whole/store"

// Main wipes the customers.
func Main() int32 {
	store.Put("wiped", "yes")
	return 1
}
