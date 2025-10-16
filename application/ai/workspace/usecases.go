// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workspace

import (
	"iter"
	"sync"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

// Platform identifies which implementation should be used.
type Platform string

func (p Platform) Identity() Platform {
	return p
}

func (p Platform) WithIdentity(o Platform) Platform {
	return o
}

func (p Platform) String() string {
	switch p {
	case MistralAI:
		return "Mistral AI"
	case OpenAI:
		return "Open AI"
	default:
		return string(p)
	}
}

const (
	MistralAI Platform = "mistralai"
	OpenAI    Platform = "openai"
)

type ID string

// Workspace describes a collective of agent also known as a multi-agent system (MAS).
// A workspace of agents is always bound to a specific platform, and it is not possible to mix agents of different
// ensembles. This is a design decision to allow cloud-based agents connected to a specific api token, thus it is
// impossible to mix cloud agents between different providers and even the same provider but with different api tokens
// (as of 2025). If you need to mix agents, you need to model and provide them as functions to specific agents.
type Workspace struct {
	ID          ID           `json:"id,omitempty"`
	Agents      []agent.ID   `json:"agents,omitempty"`
	Name        string       `json:"name,omitempty"`
	Description string       `json:"desc,omitempty"`
	Platform    Platform     `json:"platform,omitempty"`
	Libraries   []library.ID `json:"libraries,omitempty"`
	System      bool         `json:"userEditable,omitempty"`
	LastMod     time.Time    `json:"lastMod,omitempty"`
	LastModBy   user.ID      `json:"lastModBy,omitempty"`
	CreatedAt   time.Time    `json:"createdAt,omitempty"`
}

func (e Workspace) Identity() ID {
	return e.ID
}

type Repository data.Repository[Workspace, ID]

type CreateOptions struct {
	Name        string
	Description string
	Platform    Platform
	// The system flag defines this workspace as a system workspace which is not editable through the UI.
	System struct {
		Valid bool
		ID    ID
	}
}

// Create will create a new workspace as expected but
type Create func(subject auth.Subject, createOptions CreateOptions) (ID, error)

type FindAll func(subject auth.Subject) iter.Seq2[ID, error]

type FindByID func(subject auth.Subject, id ID) (option.Opt[Workspace], error)

type DeleteByID func(subject auth.Subject, id ID) error
type UseCases struct {
	Create     Create
	FindAll    FindAll
	FindByID   FindByID
	DeleteByID DeleteByID
}

func NewUseCases(repo Repository, repoAgents agent.Repository) UseCases {
	var mutex sync.Mutex
	return UseCases{
		Create:     NewCreate(&mutex, repo),
		FindAll:    NewFindAll(repo),
		FindByID:   NewFindByID(repo),
		DeleteByID: NewDeleteByID(repo, repoAgents),
	}
}
