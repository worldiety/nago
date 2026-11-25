// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import "time"

// SetLinkTransferLimit replaces linkTransferLimit for the external tests and returns the function, which restores the
// former limit. Restore it only after the server has completed all its requests, e.g. by registering the restore
// with t.Cleanup before the server is started.
func SetLinkTransferLimit(d time.Duration) (restore func()) {
	former := linkTransferLimit
	linkTransferLimit = d
	return func() { linkTransferLimit = former }
}
