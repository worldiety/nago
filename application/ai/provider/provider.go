// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package provider

import (
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/auth"
)

type ID string

// Provider is the central abstraction around various ai implementations like OpenAI or MistralAI.
type Provider interface {
	// Identity of this provider, usually based on the used secret ID.
	Identity() ID

	// Name is usually the name of the used secret.
	Name() string

	Description() string

	Models() Models

	// Libraries returns the implementation, if this Provider supports native libraries.
	Libraries() option.Opt[Libraries]

	Agents() option.Opt[Agents]

	Conversations() option.Opt[Conversations]
}

type Models interface {
	All(subject auth.Subject) iter.Seq2[model.Model, error]
}

type Libraries interface {
	Create(subject auth.Subject, opts library.CreateOptions) (library.Library, error)
	FindByID(subject auth.Subject, id library.ID) (option.Opt[library.Library], error)
	All(subject auth.Subject) iter.Seq2[library.Library, error]
	Delete(subject auth.Subject, id library.ID) error
	Update(subject auth.Subject, id library.ID, update library.UpdateOptions) (library.Library, error)
	Library(id library.ID) Library
}

type Library interface {
	Identity() library.ID
	All(subject auth.Subject) iter.Seq2[document.Document, error]
	Delete(subject auth.Subject, doc document.ID) error
	Create(subject auth.Subject, doc document.CreateOptions) (document.Document, error)
}

type Agents interface {
	All(subject auth.Subject) iter.Seq2[agent.Agent, error]
	Delete(subject auth.Subject, id agent.ID) error
	FindByID(subject auth.Subject, id agent.ID) (option.Opt[agent.Agent], error)
	FindByName(subject auth.Subject, name string) iter.Seq2[agent.Agent, error]
	Create(subject auth.Subject, options agent.CreateOptions) (agent.Agent, error)
}

type Conversations interface {
	All(subject auth.Subject) iter.Seq2[conversation.Conversation, error]
	FindByID(subject auth.Subject, id conversation.ID) (option.Opt[conversation.Conversation], error)
	Delete(subject auth.Subject, id conversation.ID) error
	// Create returns the conversation and the according request and response messages.
	Create(subject auth.Subject, opts conversation.CreateOptions) (conversation.Conversation, []message.Message, error)
	Conversation(subject auth.Subject, id conversation.ID) Conversation
}

type Conversation interface {
	Identity() conversation.ID
	All(subject auth.Subject) iter.Seq2[message.Message, error]
	Append(subject auth.Subject, opts message.AppendOptions) ([]message.Message, error)
}
