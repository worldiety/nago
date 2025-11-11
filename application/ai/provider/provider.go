// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package provider

import (
	"errors"
	"io"
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/tool"
	"go.wdy.de/nago/auth"
)

var (
	TooManyRequests = errors.New("too many requests") // TooManyRequests tells you that the rate limiter has kicked in
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

	// Tools list all possibly available parameterless build-in tools.
	// Not all combinations with all models are allowed. We could
	// include an allow list into the tool, however, that is usually not programmatically readable from provider APIs
	// and the world is still moving too fast. We just omit that and leave it as a trial and error task for
	// the user.
	Tools() Tools

	// Libraries returns the implementation, if this Provider supports native libraries.
	Libraries() option.Opt[Libraries]

	Agents() option.Opt[Agents]

	Conversations() option.Opt[Conversations]

	// Files interface to work with submitting files into the provider and reading generated files back.
	Files() option.Opt[Files]
}

type Tools interface {
	All(subject auth.Subject) iter.Seq2[tool.Tool, error]
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

	// Create submits a file which should be inserted into the library. However, that may fail for various reasons.
	// The defined errors are specified as follows:
	//   - [document.UnsupportedFormatError] if the file type is not supported, e.g. because it cannot be parsed
	Create(subject auth.Subject, doc document.CreateOptions) (document.Document, error)
	TextContentByID(subject auth.Subject, id document.ID) (option.Opt[string], error)
	StatusByID(subject auth.Subject, id document.ID) (option.Opt[document.ProcessingStatus], error)
	FindByID(subject auth.Subject, id document.ID) (option.Opt[document.Document], error)
}

type Agents interface {
	All(subject auth.Subject) iter.Seq2[agent.Agent, error]
	Delete(subject auth.Subject, id agent.ID) error
	FindByID(subject auth.Subject, id agent.ID) (option.Opt[agent.Agent], error)
	FindByName(subject auth.Subject, name string) iter.Seq2[agent.Agent, error]
	Create(subject auth.Subject, options agent.CreateOptions) (agent.Agent, error)
	Agent(id agent.ID) Agent
}

type Agent interface {
	Identity() agent.ID
	Update(subject auth.Subject, opts agent.UpdateOptions) (agent.Agent, error)
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

	// Append takes the given options and applies it for input processing. This method blocks until processing
	// has finished and returns all input messages and the generated output messages which may require further
	// processing work.
	Append(subject auth.Subject, opts message.AppendOptions) ([]message.Message, error)
}

type Files interface {
	All(subject auth.Subject) iter.Seq2[file.File, error]
	FindByID(subject auth.Subject, id file.ID) (option.Opt[file.File], error)
	Delete(subject auth.Subject, id file.ID) error
	Put(subject auth.Subject, opts file.CreateOptions) (file.File, error)
	Get(subject auth.Subject, id file.ID) (option.Opt[io.ReadCloser], error)
}
