// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package conversation

import "go.wdy.de/nago/application/permission"

var (
	PermStart   = permission.DeclareCreate[Start]("nago.ai.conversation.start", "AI Conversation")
	PermAppend  = permission.DeclareCreate[Append]("nago.ai.conversation.append", "AI Message")
	PermFindAll = permission.DeclareFindAll[FindAll]("nago.ai.conversation.findall", "AI Conversation")
	PermDelete  = permission.DeclareFindAll[Delete]("nago.ai.conversation.delete", "AI Conversation")
)
