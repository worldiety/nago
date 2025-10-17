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
}

func NewClient(token string) *Client {
	return &Client{
		c: &http.Client{
			Timeout: 60 * time.Second,
		},
		base:  "https://api.mistral.ai/v1/",
		token: token,
	}
}

func (c *Client) CreateConversion(req CreateConversionRequest) (CreateConversionResponse, error) {
	var resp CreateConversionResponse
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
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

func (c *Client) ListAgents() iter.Seq2[Agent, error] {
	return func(yield func(Agent, error) bool) {
		var resp []Agent
		err := xhttp.NewRequest().
			Client(c.c).
			BaseURL(c.base).
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
