// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package message

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
)

type ID string

// InputText is a text message usually created by a human.
type InputText struct {
	Text string
}

type Content struct {
	Text option.Ptr[string] `json:"text,omitzero"`
}

type Role string

const (
	User          Role = "user"
	AssistantRole      = "assistant"
)

type Message struct {
	ID        ID                     `json:"id"`
	CreatedAt xtime.UnixMilliseconds `json:"createdAt"`
	CreatedBy user.ID                `json:"createdBy"`

	// deprecated
	Inputs xslices.Slice[Content] `json:"inputs"`

	Role          Role               `json:"role"`
	MessageInput  option.Ptr[string] `json:"messageInput"`
	MessageOutput option.Ptr[string] `json:"messageOutput"`
}

func (m Message) Identity() ID {
	return m.ID
}

type Repository data.Repository[Message, ID]
