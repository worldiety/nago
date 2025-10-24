// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"encoding/json"
	"net/http"
	"time"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/pkg/xhttp"
)

type CreateConversionRequest struct {
	Inputs      string `json:"inputs"`
	Stream      bool   `json:"stream"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Store       bool   `json:"store,omitempty"`
	AgentId     string `json:"agent_id,omitempty"`
	Model       string `json:"model,omitempty"`

	HandoffExecution interface{} `json:"handoff_execution,omitempty"`
	Instructions     interface{} `json:"instructions,omitempty"`
	Tools            interface{} `json:"tools,omitempty"`
	CompletionArgs   interface{} `json:"completion_args,omitempty"`
}

type CreateConversionResponse struct {
	Object         string `json:"object"`
	ConversationId string `json:"conversation_id"`
	Outputs        []struct {
		Object      string    `json:"object"`
		Type        string    `json:"type"`
		CreatedAt   time.Time `json:"created_at"`
		CompletedAt time.Time `json:"completed_at"`
		Id          string    `json:"id"`
		AgentId     string    `json:"agent_id"`
		Model       string    `json:"model"`
		Role        string    `json:"role"`
		Content     string    `json:"content"`
	} `json:"outputs"`
	Usage struct {
		PromptTokens     int         `json:"prompt_tokens"`
		CompletionTokens int         `json:"completion_tokens"`
		TotalTokens      int         `json:"total_tokens"`
		ConnectorTokens  interface{} `json:"connector_tokens"`
		Connectors       interface{} `json:"connectors"`
	} `json:"usage"`
}

type Client struct {
	c     *http.Client
	token string
	base  string
	retry int
}

func NewClient(token string) *Client {
	// note, we get a lot of connection problems with the mistral api, like
	//   got 503 Service Unavailable: upstream connect error or disconnect/reset before headers.
	//   retried and the latest reset reason: remote connection failure, transport failure reason:
	//   TLS_error:|268436501:SSL routines:OPENSSL_internal:SSLV3_ALERT_CERTIFICATE_EXPIRED:TLS_error_end"
	// and it looks like they don't get their cluster stuff configured the right way

	// also note, that in general their service looks very unreliable. During development, we got a lot of random
	// http 503 service not available responses.
	return &Client{
		c: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
		base:  "https://api.mistral.ai/v1/",
		token: token,
		retry: 3,
	}
}

type Agent struct {
	Instructions string `json:"instructions"`
	Tools        []struct {
		Type     string `json:"type"`
		Function struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Strict      bool   `json:"strict"`
			Parameters  struct {
			} `json:"parameters"`
		} `json:"function"`
	} `json:"tools"`
	CompletionArgs struct {
		Stop             string  `json:"stop"`
		PresencePenalty  int     `json:"presence_penalty"`
		FrequencyPenalty int     `json:"frequency_penalty"`
		Temperature      float64 `json:"temperature"`
		TopP             int     `json:"top_p"`
		MaxTokens        int     `json:"max_tokens"`
		RandomSeed       int     `json:"random_seed"`
		Prediction       struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		} `json:"prediction"`
		ResponseFormat struct {
			Type       string `json:"type"`
			JsonSchema struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Schema      struct {
				} `json:"schema"`
				Strict bool `json:"strict"`
			} `json:"json_schema"`
		} `json:"response_format"`
		ToolChoice string `json:"tool_choice"`
	} `json:"completion_args"`
	Model       string    `json:"model"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Handoffs    []string  `json:"handoffs"`
	Object      string    `json:"object"`
	Id          string    `json:"id"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (a Agent) IntoAgent() agent.Agent {
	return agent.Agent{
		ID:          agent.ID(a.Id),
		Name:        a.Name,
		Description: a.Description,
	}
}

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

type UpdateAgentRequest struct {
	Instructions *string `json:"instructions"`
	Model        *string `json:"model"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
}

func (c *Client) UpdateAgent(id string, req UpdateAgentRequest) error {
	return xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		URL("agents/" + id).
		BodyJSON(req).
		Assert2xx(true).
		BearerAuthentication(c.token).
		Patch()
}

type AppendConversationRequest struct {
	Inputs []Input `json:"inputs,omitempty"`
	Store  bool    `json:"store"`
	Stream bool    `json:"stream"`
}

type AppendConversationResponse struct {
	ConversationId string `json:"conversation_id"`
	Outputs        []struct {
		Content string `json:"content"`
	} `json:"outputs"`
	Usage struct {
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
	ConversationId string `json:"conversation_id"`
	Outputs        []struct {
		Content string `json:"content"`
	} `json:"outputs"`
	Usage struct {
	} `json:"usage"`
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
