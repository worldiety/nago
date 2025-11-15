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
	"go.wdy.de/nago/application/chatbot/channel"
	"go.wdy.de/nago/application/chatbot/message"
	"go.wdy.de/nago/application/chatbot/user"
	"go.wdy.de/nago/auth"
)

type ID string
type Provider interface {
	Users() Users
	Channels() Channels

	Identity() ID
	Name() string
}

type Channels interface {
	Create(subject auth.Subject, users ...user.ID) (channel.Channel, error)
	Channel(id channel.ID) Channel
}

type Channel interface {
	Post(subject auth.Subject, opts message.CreateOptions) (message.Message, error)
}

type Users interface {
	Me(subject auth.Subject) (user.User, error)
	All(subject auth.Subject) iter.Seq2[user.User, error]
	FindByEmail(subject auth.Subject, mail user.Email) (option.Opt[user.User], error)
}
