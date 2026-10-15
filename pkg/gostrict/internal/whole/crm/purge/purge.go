// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package purge is a use case of the area crm, which calls a use case, which the user may not call.
package purge

import (
	"go.wdy.de/nago/pkg/gostrict/internal/whole/admin/wipe"
	"go.wdy.de/nago/pkg/gostrict/internal/whole/store"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Main marks the purge and wipes, which the aspect denies, so that the whole call is aborted.
func Main() string {
	store.Put("purge", "started")
	return "wiped " + text.FromInt(wipe.Main())
}
