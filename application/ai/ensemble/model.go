// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ensemble

import (
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/pkg/data"
)

// Platform identifies which implementation should be used.
type Platform string

const (
	MistralAI Platform = "mistralai"
	OpenAI    Platform = "openai"
)

type ID string

// Ensemble describes a collective of agent also known as a multi-agent system (MAS).
// An ensemble of agents is always bound to a specific platform, and it is not possible to mix agents of different
// ensembles. This is a design decision to allow cloud-based agents connected to a specific api token, thus it is
// impossible to mix cloud agents between different providers and even the same provider but with different api tokens
// (as of 2025). If you need to mix agents, you need to model and provide them as functions to specific agents.
type Ensemble struct {
	ID           ID         `json:"id,omitempty"`
	Agents       []agent.ID `json:"agents,omitempty"`
	Name         string     `json:"name,omitempty"`
	Description  string     `json:"desc,omitempty"`
	Platform     Platform   `json:"platform,omitempty"`
	UserEditable bool       `json:"userEditable,omitempty"`
}

func (e Ensemble) Identity() ID {
	return e.ID
}

type Repository data.Repository[Ensemble, ID]
