// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package conversation

import (
	"iter"
	"sync"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xtime"
)

type ID string

type State int

const (
	StatePending State = iota
	StateSynced
)

type Conversation struct {
	ID        ID           `json:"id,omitempty"`
	Workspace workspace.ID `json:"workspace,omitempty"`
	Agent     agent.ID     `json:"agent,omitempty"`

	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`

	// Instructions is the initial chat prompt.
	Instructions string `json:"instructions,omitempty"`

	// Input is a slice of union types of various content types. This must not be empty.
	Input []message.Content `json:"input,omitempty"`

	// CloudStore indicates if the conversation should be stored and retrievable if the provider uses a cloud
	// backend.
	CloudStore bool                   `json:"cloudStore,omitempty"`
	CreatedAt  xtime.UnixMilliseconds `json:"createdAt,omitempty"`
	CreatedBy  user.ID                `json:"createdBy,omitempty"`

	State State  `json:"state,omitempty"`
	Error string `json:"error,omitempty"`
}

func (c Conversation) Identity() ID {
	return c.ID
}

type InputEntry interface {
	inputEntry()
}

type StartOptions struct {
	Workspace     workspace.ID
	WorkspaceName string // alternative to Workspace ID find the first workspace with the given name
	Agent         agent.ID
	AgentName     string // alternative to Agent ID find the first agent with the given name

	Name        string
	Description string

	// Instructions is the initial chat prompt.
	Instructions string

	// Input is a slice of union types of various content types. This must not be empty.
	Input []message.Content

	// CloudStore indicates if the conversation should be stored and retrievable if the provider uses a cloud
	// backend.
	CloudStore bool
}

type Start func(subject auth.Subject, opts StartOptions) (ID, error)

// FindAll returns all those Conversations which are either assigned by global [PermFindAll] or individually assigned
// resource permissions or if the conversation was created by the subject.
type FindAll func(subject auth.Subject) iter.Seq2[Conversation, error]

type AppendOptions struct {
	Conversation ID
	// Input is a slice of union types of various content types. This must not be empty.
	Input []message.Content

	// CloudStore indicates if the conversation should be stored and retrievable if the provider uses a cloud
	// backend.
	CloudStore bool
}

// Append adds a message to the conversation.
type Append func(subject auth.Subject, opts AppendOptions) (message.ID, error)

type Repository data.Repository[Conversation, ID]

type UseCases struct {
	Start   Start
	Append  Append
	FindAll FindAll
}

func NewUseCases(bus events.Bus, repo Repository, repoWS workspace.Repository, repoAgents agent.Repository, repoMsg message.Repository, idxConvMsg *data.CompositeIndex[ID, message.ID]) UseCases {
	var mutex sync.Mutex
	return UseCases{
		Start:   NewStart(&mutex, bus, repo, repoWS, repoAgents),
		Append:  NewAppend(&mutex, bus, repo, repoMsg, idxConvMsg),
		FindAll: NewFindAll(repo),
	}
}
