// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uievs

import "go.wdy.de/nago/presentation/core"

type Pages struct {
	Audit  core.NavigationPath // Path to the "audit" page, e.g. "/admin/events/<prefix>/audit"
	Create core.NavigationPath // Path to the "create" page, e.g. "/admin/events/<prefix>/create/<discriminator>"
	Index  core.NavigationPath // Path to the "index" page, e.g. "/admin/events/<prefix>/index/<index-id>"
}
