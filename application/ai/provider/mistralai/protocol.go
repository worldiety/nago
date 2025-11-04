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
			Timeout: 120 * time.Second, // mistral can become enormously slow, thus let us try a 2-minute timeout
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
		base:  "https://api.mistral.ai/v1/",
		token: token,
		retry: 3,
	}
}
