// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ucrebac

import "go.wdy.de/nago/application/permission"

// note: we need this unwanted package because we integrated the rebac package into the application scope which causes a dependency cycle for any regular permission case
// DO NOT re-use this pattern for regular modeling: this is only due to bootstrapping.

var (
	PermFindAllResources = permission.DeclareFindAll[FindAllResources]("nago.rebac.resources.find_all", "Resources")
	PermWithReBAC        = permission.DeclareUpdate[WithReBAC]("nago.rebac.resources.with_db", "Resources")
)
