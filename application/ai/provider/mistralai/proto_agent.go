// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"iter"
	"time"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/pkg/xhttp"
	"go.wdy.de/nago/pkg/xtime"
)

type CreateAgentRequest struct {
	Model        string   `json:"model"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Handoffs     []string `json:"handoffs,omitempty"`
	Instructions string   `json:"instructions,omitempty"`
}

type AgentInfo struct {
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

func (a AgentInfo) IntoAgent() agent.Agent {
	return agent.Agent{
		ID:          agent.ID(a.Id),
		Name:        a.Name,
		Description: a.Description,
		Prompt:      a.Instructions,
		Model2:      model.ID(a.Model),
		LastMod:     xtime.UnixMilliseconds(a.UpdatedAt.UnixMilli()),
	}
}

func (c *Client) CreateAgent(req CreateAgentRequest) (AgentInfo, error) {
	var resp AgentInfo
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		URL("agents").
		BearerAuthentication(c.token).
		Assert2xx(true).
		BodyJSON(req).
		ToJSON(&resp).
		Post()

	return resp, err
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
	return xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		Assert2xx(true).
		URL("agents/" + id).
		BearerAuthentication(c.token).
		Delete()
}
