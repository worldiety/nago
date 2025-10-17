// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package agent

import "go.wdy.de/nago/application/permission"

var (
	PermFindByID = permission.DeclareFindByID[FindByID]("nago.ai.agent.find_by_id", "AI Agent")
	PermUpdate   = permission.DeclareUpdate[Update]("nago.ai.agent.update", "AI Agent")
)
