// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cache

import "go.wdy.de/nago/application/permission"

type Stub func()

// Libraries Permissions

var (
	PermLibraryFindByID = permission.DeclareFindByID[Stub]("nago.ai.library.find_by_id", "AI Library")
	PermLibraryFindAll  = permission.DeclareFindAll[Stub]("nago.ai.library.find_all", "AI Library")
	PermLibraryCreate   = permission.DeclareCreate[Stub]("nago.ai.library.create", "AI Library")
	PermLibraryDelete   = permission.DeclareDeleteByID[Stub]("nago.ai.library.delete", "AI Library")
	PermLibraryUpdate   = permission.DeclareUpdate[Stub]("nago.ai.library.update", "AI Library")
)

// Document Permissions

var (
	PermDocumentFindAll = permission.DeclareFindAll[Stub]("nago.ai.document.find_all", "AI Document")
	PermDocumentDelete  = permission.DeclareDeleteByID[Stub]("nago.ai.document.delete", "AI Document")
	PermDocumentCreate  = permission.DeclareCreate[Stub]("nago.ai.document.create", "AI Document")
)

// Agent Permissions

var (
	PermAgentFindAll    = permission.DeclareFindAll[Stub]("nago.ai.agent.find_all", "AI Agent")
	PermAgentDelete     = permission.DeclareDeleteByID[Stub]("nago.ai.agent.delete", "AI Agent")
	PermAgentCreate     = permission.DeclareCreate[Stub]("nago.ai.agent.create", "AI Agent")
	PermAgentFindByID   = permission.DeclareFindByID[Stub]("nago.ai.agent.find_by_id", "AI Agent")
	PermAgentFindByName = permission.DeclareFindByName[Stub]("nago.ai.agent.find_by_name", "AI Agent")
	PermAgentUpdate     = permission.DeclareUpdate[Stub]("nago.ai.agent.update", "AI Agent")
)

// Conversations Permissions

var (
	PermConversationFindAll  = permission.DeclareFindAll[Stub]("nago.ai.conversation.find_all", "AI Conversation")
	PermConversationFindByID = permission.DeclareFindByID[Stub]("nago.ai.conversation.find_by_id", "AI Conversation")
	PermConversationDelete   = permission.DeclareDeleteByID[Stub]("nago.ai.conversation.delete", "AI Conversation")
	PermConversationCreate   = permission.DeclareCreate[Stub]("nago.ai.conversation.create", "AI Conversation")
)

// Message Permission

var (
	PermMessageFindAll = permission.DeclareFindAll[Stub]("nago.ai.message.find_all", "AI Message")
	PermMessageAppend  = permission.DeclareAppend[Stub]("nago.ai.message.append", "AI Message")
)

// Model Permissions

var (
	PermFindAllModel = permission.DeclareFindAll[Stub]("nago.ai.model.find_all", "AI Model")
)
