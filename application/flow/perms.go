// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import "go.wdy.de/nago/application/permission"

var (
	PermFindWorkspaces  = permission.DeclareFindAll[FindWorkspaces]("nago.flowmod.workspace.findall", "Flow Workspace")
	PermDeleteWorkspace = permission.DeclareDeleteByID[DeleteWorkspace]("nago.flowmod.workspace.delete", "Flow Workspace")
	PermExportWorkspace = permission.DeclareExportByID[ExportWorkspace]("nago.flowmod.workspace.export", "Flow Workspace")
	PermImportWorkspace = permission.DeclareImportByID[ImportWorkspace]("nago.flowmod.workspace.import", "Flow Workspace")
)
