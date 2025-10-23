// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ai

import "go.wdy.de/nago/application/permission"

var (
	PermFindProviderByName = permission.DeclareFindByName[FindProviderByName]("nago.ai.provider.find_by_name", "AI Provider")
	PermFindProviderByID   = permission.DeclareFindByID[FindProviderByID]("nago.ai.provider.find_by_id", "AI Provider")
	PermFindAllProvider    = permission.DeclareFindAll[FindAllProvider]("nago.ai.provider.find_all", "AI Provider")
)
