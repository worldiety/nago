// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workspace

import "go.wdy.de/nago/application/permission"

var (
	PermCreate     = permission.DeclareCreate[Create]("nago.ai.workspace.create", "AI Workspace")
	PermFindByID   = permission.DeclareFindByID[FindByID]("nago.ai.workspace.find_by_id", "AI Workspace")
	PermFindAll    = permission.DeclareFindAllIdentifiers[FindAll]("nago.ai.workspace.find_all", "AI Workspace")
	PermDeleteByID = permission.DeclareDeleteByID[DeleteByID]("nago.ai.workspace.delete", "AI Workspace")

	PermCreateAgent = permission.DeclareCreate[CreateAgent]("nago.ai.workspace.create_agent", "AI Agent")
	PermDeleteAgent = permission.DeclareDeleteByID[DeleteAgent]("nago.ai.workspace.delete_agent", "AI Agent")
)
