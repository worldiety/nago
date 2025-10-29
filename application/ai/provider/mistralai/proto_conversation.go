// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"encoding/json"
	"time"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/pkg/xhttp"
	"go.wdy.de/nago/pkg/xtime"
)

func (c *Client) DeleteConversation(id string) error {
	return xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		Assert2xx(true).
		URL("conversations/" + id).
		BearerAuthentication(c.token).
		Delete()
}

type AppendConversationRequest struct {
	Inputs []Input `json:"inputs,omitempty"`
	Store  bool    `json:"store"`
	Stream bool    `json:"stream"`
}

type AppendConversationResponse struct {
	ConversationId string     `json:"conversation_id"`
	Outputs        []EntryBox `json:"outputs"`
	Usage          struct {
		CompletionTokens int `json:"completion_tokens"`
		PromptTokens     int `json:"prompt_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *Client) AppendConversation(conversationId string, req AppendConversationRequest) (AppendConversationResponse, error) {
	var resp AppendConversationResponse
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		URL("conversations/" + conversationId).
		Assert2xx(true).
		BearerAuthentication(c.token).
		BodyJSON(req).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Post()

	return resp, err
}

type CreateConversationRequest struct {
	AgentID      string  `json:"agent_id,omitempty"`
	Description  string  `json:"description,omitempty"`
	Name         string  `json:"name,omitempty"`
	Instructions string  `json:"instructions,omitempty"`
	Model        string  `json:"model,omitempty"`
	Store        bool    `json:"store"`
	Stream       bool    `json:"stream"`
	Inputs       []Input `json:"inputs,omitempty"`
}

type CreateConversationResponse struct {
	ConversationId string     `json:"conversation_id"`
	Object         string     `json:"object"`
	Outputs        []EntryBox `json:"outputs"`
	Usage          struct {
		CompletionTokens int `json:"completion_tokens"`
		PromptTokens     int `json:"prompt_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (c CreateConversationResponse) IntoConversation() conversation.Conversation {
	return conversation.Conversation{
		ID: conversation.ID(c.ConversationId),
	}
}

func (c *Client) CreateConversation(req CreateConversationRequest) (CreateConversationResponse, error) {
	var resp CreateConversationResponse
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		URL("conversations").
		Assert2xx(true).
		BearerAuthentication(c.token).
		BodyJSON(req).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Post()

	return resp, err
}

type ConversationInfo struct {
	AgentId      string    `json:"agent_id"`
	CreatedAt    time.Time `json:"created_at"`
	Description  string    `json:"description"`
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	Object       string    `json:"object"`
	UpdatedAt    time.Time `json:"updated_at"`
	Model        string    `json:"model"`
	Instructions string
}

func (c ConversationInfo) IntoConversation() conversation.Conversation {
	return conversation.Conversation{
		ID:           conversation.ID(c.Id),
		Agent:        agent.ID(c.AgentId),
		Name:         c.Name,
		Description:  c.Description,
		Instructions: c.Instructions,
		CreatedAt:    xtime.UnixMilliseconds(c.CreatedAt.UnixMilli()),
		CreatedBy:    "", //?
	}
}

func (c *Client) ListConversations() ([]ConversationInfo, error) {
	var resp []ConversationInfo
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		Query("page", "0").
		Query("page_size", "100000").
		URL("conversations").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp, err
}

func (c *Client) GetConversation(convId string) (ConversationInfo, error) {
	var resp ConversationInfo
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		URL("conversations/" + convId).
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp, err
}

type Input interface {
	Type() string
	MarshalJSON() ([]byte, error)
	isInput()
}

type Role string

const (
	RoleAssistant Role = "assistant"
	RoleUser      Role = "user"
)

type MessageInputEntry struct {
	Content Chunk  `json:"content,omitempty"`
	Role    Role   `json:"role,omitempty"`
	ID      string `json:"id,omitempty"`
}

func (e MessageInputEntry) MarshalJSON() ([]byte, error) {
	type jsonInputEntry struct {
		Content json.RawMessage `json:"content,omitempty"`
		Role    Role            `json:"role,omitempty"`
		ID      string          `json:"id,omitempty"`
		Type    string          `json:"type,omitempty"`
		Object  string          `json:"object,omitempty"`
	}

	raw, err := json.Marshal(e.Content)
	if err != nil {
		return nil, err
	}

	return json.Marshal(jsonInputEntry{
		Content: raw,
		Role:    e.Role,
		ID:      e.ID,
		Type:    e.Type(),
		Object:  "entry",
	})
}

func (e MessageInputEntry) isInput() {}

func (e MessageInputEntry) Type() string {
	return "message.input"
}

type Chunk interface {
	Type() string
	MarshalJSON() ([]byte, error)
	isChunk()
}

type TextChunk struct {
	Text string `json:"text"`
}

func (t TextChunk) Type() string {
	return "text"
}

func (t TextChunk) isChunk() {}

func (t TextChunk) MarshalJSON() ([]byte, error) {
	const mistralApiDocIsJustWrongAgain = true
	if mistralApiDocIsJustWrongAgain {
		// TODO this sucks as hell, the api doc clearly tells us how a TextChunk must be encoded but the server rejects with 422, no extra inputs permitted
		// TODO if you look at https://docs.mistral.ai/agents/agents you can see that the used encoding was never defined in the api docs and does not conform to the Chunk spec.
		return json.Marshal(t.Text)
	}

	type jsonTextChunk struct {
		Text string `json:"text"`
		Type string `json:"type"`
	}

	return json.Marshal(jsonTextChunk{
		Text: t.Text,
		Type: t.Type(),
	})
}
