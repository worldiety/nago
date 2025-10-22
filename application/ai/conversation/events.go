// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package conversation

import (
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/user"
)

type Started struct {
	Conversation ID
	ByUser       user.ID
	Error        string
}
type SyncStatusUpdated struct {
	Conversation ID
	ByUser       user.ID
	Error        string
}

type MessageAppended struct {
	Conversation ID
	Message      message.ID
}

type AgentAppended struct {
	Conversation ID
	Content      []message.Content
}

type HumanAppended struct {
	Conversation ID
	ByUser       user.ID
	Content      []message.Content
}

type Updated struct {
	Conversation ID
}
