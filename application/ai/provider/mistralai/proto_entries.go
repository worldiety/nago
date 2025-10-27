// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/pkg/xhttp"
	"go.wdy.de/nago/pkg/xtime"
)

type ListEntryResponse struct {
	ConversationId string     `json:"conversation_id"`
	Object         string     `json:"object"` // e.g. "conversation.history",
	Entries        []EntryBox `json:"entries"`
}

func (c *Client) ListEntries(conversationId string) ([]EntryBox, error) {
	var resp ListEntryResponse
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		Query("page", "0").
		Query("page_size", "100000").
		URL("conversations/" + conversationId + "/history").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp.Entries, err
}

type EntryBox struct {
	Object string `json:"object"` // "object":"entry"
	Type   string `json:"type"`   // "type":"message.input"|"message.output"
	Value  Entry  `json:"-"`
}

func (e EntryBox) MarshalJSON() ([]byte, error) {
	switch m := e.Value.(type) {
	case MessageInput:
		m.Object = "entry"
		m.Type = "message.input"
		return json.Marshal(m)
	case MessageOutput:
		m.Object = "entry"
		m.Type = "message.output"
		return json.Marshal(m)
	default:
		return nil, fmt.Errorf("unknown entry box type: %T", e.Value)
	}
}

func (e *EntryBox) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Object string `json:"object"` // "object":"entry"
		Type   string `json:"type"`   // "type":"message.input"|"message.output"
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp.Object != "entry" {
		return fmt.Errorf("unknown entry box object type: %s", e.Object)
	}

	switch tmp.Type {
	case "message.input":
		var tmp MessageInput
		if err := json.Unmarshal(data, &tmp); err != nil {
			return err
		}

		e.Value = tmp
		return nil
	case "message.output":
		var tmp MessageOutput
		if err := json.Unmarshal(data, &tmp); err != nil {
			return err
		}
		e.Value = tmp
		return nil
	default:
		return fmt.Errorf("unknown entry box type: %s: %s", e.Type, string(data))
	}
}

type Entry interface {
	isEntry()
	IntoMessage() message.Message
}

type MessageInput struct {
	Object string `json:"object"` // "object":"entry"
	Type   string `json:"type"`   // "type":"message.input"
	Role   string `json:"role"`   // e.g. "role":"user"

	Id          string      `json:"id"`
	CompletedAt interface{} `json:"completed_at"`
	Content     string      `json:"content"`
	CreatedAt   time.Time   `json:"created_at"`
	Prefix      bool        `json:"prefix"`
}

func (e MessageInput) isEntry() {}

func (e MessageInput) IntoMessage() message.Message {
	return message.Message{
		ID:           message.ID(e.Id),
		CreatedAt:    xtime.UnixMilliseconds(e.CreatedAt.UnixMilli()),
		CreatedBy:    "", //todo ??
		Role:         message.Role(e.Role),
		MessageInput: option.Pointer(&e.Content),
	}
}

type MessageOutput struct {
	Object string `json:"object"` // "object":"entry"
	Type   string `json:"type"`   // "type":"message.output"
	Role   string `json:"role"`   // e.g. "role":"assistant"

	Id          string      `json:"id"`
	AgentId     string      `json:"agent_id"`
	CompletedAt interface{} `json:"completed_at"`
	Content     string      `json:"content"`
	CreatedAt   time.Time   `json:"created_at"`
	Model       string      `json:"model"`
}

func (e MessageOutput) isEntry() {}

func (e MessageOutput) IntoMessage() message.Message {
	return message.Message{
		ID:            message.ID(e.Id),
		CreatedAt:     xtime.UnixMilliseconds(e.CreatedAt.UnixMilli()),
		CreatedBy:     "", //todo ??
		Role:          message.Role(e.Role),
		MessageOutput: option.Pointer(&e.Content),
	}
}
