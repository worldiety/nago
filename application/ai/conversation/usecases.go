// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package conversation

import (
	"iter"
	"log/slog"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/eventstore"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
)

type ID string

type State int

const (
	StatePending State = iota
	StateSynced
)

type Conversation struct {
	ID ID `json:"id,omitempty"`

	Agent agent.ID `json:"agent,omitempty"`
	Model model.ID `json:"model,omitempty"` // alternatively a model was used instead of an agent

	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`

	// Instructions is the initial chat prompt.
	Instructions string `json:"instructions,omitempty"`

	// Input is a slice of union types of various content types. This must not be empty.
	Input []message.Content `json:"input,omitempty"`

	// CloudStore indicates if the conversation should be stored and retrievable if the provider uses a cloud
	// backend.
	// deprecated
	CloudStore bool                   `json:"cloudStore,omitempty"`
	CreatedAt  xtime.UnixMilliseconds `json:"createdAt,omitempty"`
	CreatedBy  user.ID                `json:"createdBy,omitempty"`

	// deprecated
	Workspace workspace.ID `json:"workspace,omitempty"`

	// deprecated
	State State `json:"state,omitempty"`
	// deprecated
	Error string `json:"error,omitempty"`
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

type StartOptions struct {
	Agent     agent.ID
	AgentName string // alternative to Agent ID find the first agent with the given name

	Name        string
	Description string

	// Instructions is the initial chat prompt.
	Instructions string

	// Input is a slice of union types of various content types. This must not be empty.
	Input []message.Content

	// CloudStore indicates if the conversation should be stored and retrievable if the provider uses a cloud
	// backend.
	CloudStore bool

	// deprecated
	Workspace workspace.ID
	// deprecated
	WorkspaceName string // alternative to Workspace ID find the first workspace with the given name
}

type Start func(subject auth.Subject, opts StartOptions) (ID, error)

// FindAll returns all those Conversations which are either assigned by global [PermFindAll] or individually assigned
// resource permissions or if the conversation was created by the subject.
type FindAll func(subject auth.Subject) iter.Seq2[Conversation, error]

// FindByID applies the same rules as [FindAll].
type FindByID func(subject auth.Subject, id ID) (option.Opt[Conversation], error)

// FindMessages returns a sequence of all associated messages in the order from oldest to newest.
// The same permission rules are applied, as for [FindAll]. If someone can find a conversation or is the owner
// he/she can also read all messages.
type FindMessages func(subject auth.Subject, cid ID) iter.Seq2[message.Message, error]

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

type Delete func(subject auth.Subject, id ID) error

type Repository data.Repository[Conversation, ID]

type UseCases struct {
	Start        Start
	Append       Append
	FindAll      FindAll
	FindMessages FindMessages
	FindByID     FindByID
	Delete       Delete
}

func NewUseCases(bus events.Bus, repo Repository, repoWS workspace.Repository, repoAgents agent.Repository, repoMsg message.Repository, idxConvMsg *data.CompositeIndex[ID, message.ID]) UseCases {
	var mutex sync.Mutex

	events.SubscribeFor(bus, func(evt AgentAppended) {
		msg := message.Message{
			ID:        message.ID(eventstore.NewID()),
			CreatedAt: xtime.Now(),
			Inputs:    xslices.Wrap(evt.Content...),
		}

		if err := repoMsg.Save(msg); err != nil {
			slog.Error("failed to save message triggered by AgentAppended", "err", err.Error())
			return
		}

		if err := idxConvMsg.Put(evt.Conversation, msg.ID); err != nil {
			slog.Error("failed to put idxConvMsg", "err", err.Error())
			return
		}

		bus.Publish(Updated{Conversation: evt.Conversation})
	})

	return UseCases{
		Start:        NewStart(&mutex, bus, repo, repoWS, repoAgents, repoMsg, idxConvMsg),
		Append:       NewAppend(&mutex, bus, repo, repoMsg, idxConvMsg),
		FindAll:      NewFindAll(repo),
		FindMessages: NewFindMessages(repo, repoMsg, idxConvMsg),
		FindByID:     NewFindByID(repo),
		Delete:       NewDelete(bus, repo, repoMsg, idxConvMsg),
	}
}
