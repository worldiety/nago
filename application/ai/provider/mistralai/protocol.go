// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
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
