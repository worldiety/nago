// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"iter"
	"net/http"
	"time"

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

func (c *Client) CreateConversion(req CreateConversionRequest) (CreateConversionResponse, error) {
	var resp CreateConversionResponse
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		URL("conversations").
		BearerAuthentication(c.token).
		BodyJSON(req).
		ToJSON(&resp).
		Post()

	return resp, err
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

func (c *Client) GetAgent(id string) (Agent, error) {
	var resp Agent
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		URL("agents/"+id).
		BearerAuthentication(c.token).
		Query("page", "0").
		Query("page_size", "1000").
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp, err
}

func (c *Client) ListAgents() iter.Seq2[Agent, error] {
	return func(yield func(Agent, error) bool) {
		var resp []Agent
		err := xhttp.NewRequest().
			Client(c.c).
			BaseURL(c.base).
			Retry(c.retry).
			URL("agents").
			BearerAuthentication(c.token).
			Query("page", "0").
			Query("page_size", "1000").
			ToJSON(&resp).
			ToLimit(1024 * 1024).
			Get()

		if err != nil {
			yield(Agent{}, err)
			return
		}

		for _, agent := range resp {
			if !yield(agent, nil) {
				return
			}
		}
	}
}

func (c *Client) DeleteAgent(id string) error {
	// note: they forgot to implement the endpoint, heck, even their UI cannot delete agents right now and gets 404
	return xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		Assert2xx(true).
		URL("agents/" + id).
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
