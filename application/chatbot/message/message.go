// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package message

import (
	"go.wdy.de/nago/application/chatbot/channel"
	"go.wdy.de/nago/pkg/data"
)

type CreateOptions struct {
	Message string
}

type ID string

type Message struct {
	ID      ID         `json:"id"`
	Channel channel.ID `json:"chan"`
	Message string     `json:"msg"`
}

func (m Message) Identity() ID {
	return m.ID
}

type Repository data.Repository[Message, ID]
