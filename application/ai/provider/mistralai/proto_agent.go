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
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/pkg/xhttp"
	"go.wdy.de/nago/pkg/xtime"
)

type ToolType string

const (
	ToolFunction         ToolType = "function"
	ToolWebSearch        ToolType = "web_search"
	ToolWebSearchPremium ToolType = "web_search_premium"
	ToolCodeInterpreter  ToolType = "code_interpreter"
	ToolImageGeneration  ToolType = "image_generation"
	ToolDocumentLibrary  ToolType = "document_library"
)

type Tool struct {
	Type ToolType `json:"type"`

	// Only valid if Type == document_library
	Libraries []library.ID `json:"library_ids,omitempty"`

	// Only valid if Type == function
	Function any `json:"function,omitempty"`
}

type CreateAgentRequest struct {
	Model        string   `json:"model"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Handoffs     []string `json:"handoffs,omitempty"`
	Instructions string   `json:"instructions,omitempty"`
}

type CompletionArgs struct {
	Stop             *string  `json:"stop"`
	PresencePenalty  *int     `json:"presence_penalty"`
	FrequencyPenalty *int     `json:"frequency_penalty"`
	Temperature      *float64 `json:"temperature"`
	TopP             *int     `json:"top_p"`
	MaxTokens        *int     `json:"max_tokens"`
	RandomSeed       *int     `json:"random_seed"`
	Prediction       *struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	} `json:"prediction"`
	ResponseFormat *struct {
		Type       string `json:"type"`
		JsonSchema struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Schema      struct {
			} `json:"schema"`
			Strict bool `json:"strict"`
		} `json:"json_schema"`
	} `json:"response_format"`
	ToolChoice *string `json:"tool_choice"`
}

type AgentInfo struct {
	Instructions   string         `json:"instructions"`
	Tools          []Tool         `json:"tools"`
	CompletionArgs CompletionArgs `json:"completion_args"`
	Model          string         `json:"model"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Handoffs       []string       `json:"handoffs"`
	Object         string         `json:"object"`
	Id             string         `json:"id"`
	Version        int            `json:"version"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (a AgentInfo) IntoAgent() agent.Agent {
	var tmp agent.Temperature
	if a.CompletionArgs.Temperature != nil {
		tmp = agent.Temperature(*a.CompletionArgs.Temperature)
	}

	ag := agent.Agent{
		ID:           agent.ID(a.Id),
		Name:         a.Name,
		Description:  a.Description,
		CreatedAt:    xtime.UnixMilliseconds(a.CreatedAt.UnixMilli()),
		UpdatedAt:    xtime.UnixMilliseconds(a.UpdatedAt.UnixMilli()),
		Model:        model.ID(a.Model),
		Temperature:  tmp,
		Instructions: a.Instructions,
	}

	for _, tool := range a.Tools {
		if tool.Type == ToolDocumentLibrary {
			ag.Libraries = tool.Libraries
		}
	}

	return ag
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

func (c *Client) GetAgent(id string) (AgentInfo, error) {
	var resp AgentInfo
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

func (c *Client) ListAgents() iter.Seq2[AgentInfo, error] {
	return func(yield func(AgentInfo, error) bool) {
		var resp []AgentInfo
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
			yield(AgentInfo{}, err)
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

type UpdateAgentRequest struct {
	Instructions   *string         `json:"instructions"`
	Model          *string         `json:"model"`
	Name           *string         `json:"name"`
	Description    *string         `json:"description"`
	Tools          []Tool          `json:"tools"`
	CompletionArgs *CompletionArgs `json:"completion_args"`
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
