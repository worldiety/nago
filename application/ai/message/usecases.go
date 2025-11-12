// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package message

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/tool"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
)

type ID string

// Input is a content usually created by a human.
type Input struct {
	Text option.Opt[string]    `json:"text,omitzero"`
	File option.Opt[file.File] `json:"file,omitzero"`
	URL  option.Opt[core.URI]  `json:"url,omitzero"`
}

type Role string

const (
	User          Role = "user"
	AssistantRole      = "assistant"
)

type AppendOptions struct {
	Role       Role
	Input      []Input
	CloudStore bool
}

type Message struct {
	ID        ID                     `json:"id"`
	CreatedAt xtime.UnixMilliseconds `json:"createdAt"`
	CreatedBy user.ID                `json:"createdBy"`

	Role          Role                      `json:"role"`
	MessageInput  option.Ptr[string]        `json:"messageInput,omitzero"`
	MessageOutput option.Ptr[string]        `json:"messageOutput,omitzero"`
	ToolExecution option.Ptr[ToolExecution] `json:"toolExecution,omitzero"`
	DocumentURL   option.Ptr[DocumentURL]   `json:"documentUrl,omitzero"`
	Reference     option.Ptr[Reference]     `json:"ref,omitzero"`
	File          option.Ptr[file.File]     `json:"file"`
}

func (m Message) Identity() ID {
	return m.ID
}

type Repository data.Repository[Message, ID]

type ToolExecution struct {
	Type      string `json:"type,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type DocumentURL struct {
	Name string   `json:"name,omitempty"`
	URL  core.URI `json:"url,omitempty"`
}

type Reference struct {
	Tool        tool.ID  `json:"tool,omitempty"`
	Title       string   `json:"title,omitempty"`
	URL         core.URI `json:"url,omitempty"`
	Description string   `json:"description,omitempty"`
}
