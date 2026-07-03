// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import "go.wdy.de/nago/application/permission"

var (
	PermCreate   = permission.DeclareCreate[Create]("nago.ai.session.create", "AI Session")
	PermFindByID = permission.DeclareFindByID[FindByID]("nago.ai.session.find_by_id", "AI Session")
	PermFindAll  = permission.DeclareFindAll[FindAll]("nago.ai.session.find_all", "AI Session")
	PermAppend   = permission.DeclareAppend[Append]("nago.ai.session.append", "AI Session")
	PermRename   = permission.DeclareUpdate[Rename]("nago.ai.session.rename", "AI Session")
	PermDelete   = permission.DeclareDeleteByID[Delete]("nago.ai.session.delete", "AI Session")
)
