// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package conversation

import (
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xtime"
)

type ID string

type Conversation struct {
	ID ID `json:"id,omitempty"`

	Agent agent.ID `json:"agent,omitempty"`
	Model model.ID `json:"model,omitempty"` // alternatively a model was used instead of an agent

	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`

	// Instructions is the initial chat prompt. This may not be combined with Agent.
	Instructions string `json:"instructions,omitempty"`

	CreatedAt xtime.UnixMilliseconds `json:"createdAt,omitempty"`
	CreatedBy user.ID                `json:"createdBy,omitempty"`
}

type CreateOptions struct {
	Model        model.ID // either Model or Agent must be used - if agents are supported at all
	Agent        agent.ID
	Name         string
	Description  string
	Instructions string
	// Input is a slice of union types of various content types. This must not be empty.
	Input []message.Content `json:"input,omitempty"`
	// CloudStore indicates if the conversation should be stored and retrievable if the provider uses a cloud
	// backend.
	CloudStore bool
}

func (c Conversation) Identity() ID {
	return c.ID
}

type InputEntry interface {
	inputEntry()
}

type Repository data.Repository[Conversation, ID]
