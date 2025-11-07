// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"go.wdy.de/nago/application/ai/model"
	"golang.org/x/text/language"
)

type ModelInfoResponse struct {
	Id           string `json:"id"`
	Capabilities struct {
		CompletionChat  bool `json:"completion_chat"`
		CompletionFim   bool `json:"completion_fim"`
		FunctionCalling bool `json:"function_calling"`
		FineTuning      bool `json:"fine_tuning"`
		Vision          bool `json:"vision"`
		Classification  bool `json:"classification"`
	} `json:"capabilities"`
	Job                         string        `json:"job"`
	Root                        string        `json:"root"`
	Object                      string        `json:"object"`
	Created                     int           `json:"created"`
	OwnedBy                     string        `json:"owned_by"`
	Name                        string        `json:"name"`
	Description                 string        `json:"description"`
	MaxContextLength            int           `json:"max_context_length"`
	Aliases                     []interface{} `json:"aliases"`
	Deprecation                 interface{}   `json:"deprecation"`
	DeprecationReplacementModel interface{}   `json:"deprecation_replacement_model"`
	DefaultModelTemperature     float64       `json:"default_model_temperature"`
	TYPE                        string        `json:"TYPE"`
	Archived                    bool          `json:"archived"`
}

func (m ModelInfoResponse) IntoModel() model.Model {
	return model.Model{
		ID:                 model.ID(m.Id),
		Name:               m.Name,
		Description:        m.Description,
		DefaultTemperature: m.DefaultModelTemperature,
	}
}

func (c *Client) ListAllModels(tag language.Tag) ([]ModelInfoResponse, error) {
	var resp struct {
		Data []ModelInfoResponse `json:"data"`
	}

	err := c.newReq().
		Header("Accept-Language", tag.String()).
		URL("models").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp.Data, err
}
