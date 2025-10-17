// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"time"

	"go.wdy.de/nago/pkg/xhttp"
)

type CreateAgentRequest struct {
	Model        string   `json:"model"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Handoffs     []string `json:"handoffs,omitempty"`
	Instructions string   `json:"instructions,omitempty"`
}

type CreateAgentResponse struct {
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
		Stop             string `json:"stop"`
		PresencePenalty  int    `json:"presence_penalty"`
		FrequencyPenalty int    `json:"frequency_penalty"`
		Temperature      int    `json:"temperature"`
		TopP             int    `json:"top_p"`
		MaxTokens        int    `json:"max_tokens"`
		RandomSeed       int    `json:"random_seed"`
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

func (c *Client) CreateAgent(req CreateAgentRequest) (CreateAgentResponse, error) {
	var resp CreateAgentResponse
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		URL("agents").
		BearerAuthentication(c.token).
		BodyJSON(req).
		ToJSON(&resp).
		Post()

	return resp, err
}
