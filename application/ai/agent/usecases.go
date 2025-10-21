// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package agent

import (
	"log/slog"
	"sync"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
)

// Model represents quality categories of agent models to use. Currently, three classifications are defined
// and an actual implementation must be able to match these. However, to peek through the abstraction and pick
// a specific model, implementations may also support specific model unportable identifiers.
type Model string

func (c Model) WithIdentity(id Model) Model {
	return id
}

func (c Model) Identity() Model {
	return c
}

var Models = xslices.Wrap(Quality, Balanced, Efficiency)

const (
	// Quality refers to the most expensive and most complex reasoning models and may map in 2025 to
	// openai GPT-5 or comparable.
	Quality Model = "quality"

	// Balanced is the mid-tier which trade-off latency and cost. In 2025 this may map to OpenAI GPT-5 mini
	// or comparable.
	Balanced Model = "balanced"

	// Efficiency is the low-end which is the fastest and cheapest model. In 2025 this may map to OpenAI GPT-5 nano.
	Efficiency Model = "efficiency"
)

// Capability hints additional intentions which the agent should solve. This allows further classification of
// the general purpose Model specifier. This helps a specific implementation to pick a better concrete model.
type Capability string

func (c Capability) WithIdentity(id Capability) Capability {
	return id
}

func (c Capability) Identity() Capability {
	return c
}

func (c Capability) String() string {
	return string(c)
}

var Capabilities = xslices.Wrap(Reasoning, Coding, Multimodal, OCR, Voice, Classifier, Embedding)

const (
	Reasoning  Capability = "reasoning"
	Coding     Capability = "coding"
	Multimodal Capability = "multimodal"
	OCR        Capability = "ocr"
	Voice      Capability = "voice"
	Classifier Capability = "classifier"
	Embedding  Capability = "embedding"
)

type ID string

type State int

const (
	StatePending State = iota
	StateSynced
)

// Temperature defines how reproducible the answers of the Agent are. Lower values make the results more predictable.
// Higher values make the results more novel. The allowed range is [0..1].
type Temperature float64
type Agent struct {
	ID           ID           `json:"id,omitempty"`
	Name         string       `json:"name,omitempty"`
	Description  string       `json:"desc,omitempty"`
	Prompt       string       `json:"prompt,omitempty"`
	Model        Model        `json:"model,omitempty"`
	Libraries    []library.ID `json:"libraries,omitempty"`
	Capabilities []Capability `json:"capabilities,omitempty"`
	Temperature  Temperature  `json:"tmp,omitempty"`
	// System indicates that any modifications through the UI are not allowed. The advantage is, that the truth
	// of this agent definition is always the source code and does not contain any user modifications.
	System  bool                   `json:"system,omitempty"`
	LastMod xtime.UnixMilliseconds `json:"lastMod,omitempty"`

	State State  `json:"state,omitempty"`
	Error string `json:"error,omitempty"`
}

func (e Agent) Identity() ID {
	return e.ID
}

type Repository data.Repository[Agent, ID]

type FindByID func(subject auth.Subject, id ID) (option.Opt[Agent], error)

type Update func(subject auth.Subject, ag Agent) error

// ProvideAgent declares a system agent using a pre-defined non-empty identifier.
type ProvideAgent func(agent Agent) error

type UseCases struct {
	FindByID     FindByID
	Update       Update
	ProvideAgent ProvideAgent
}

func NewUseCases(bus events.Bus, repo Repository) UseCases {
	var mutex sync.Mutex

	events.SubscribeFor[SyncStatusUpdated](bus, func(evt SyncStatusUpdated) {
		mutex.Lock()
		defer mutex.Unlock()
		optAgent, err := repo.FindByID(evt.Agent)
		if err != nil {
			slog.Error("cannot update agent sync status: failed to find agent", "agent", evt.Agent, "err", err.Error())
			return
		}

		if optAgent.IsNone() {
			slog.Error("cannot update agent sync status: agent is gone", "agent", evt.Agent)
			return
		}

		agent := optAgent.Unwrap()
		agent.Error = evt.Error
		if agent.Error == "" {
			agent.State = StateSynced
		} else {
			agent.State = StatePending
		}

		if err := repo.Save(agent); err != nil {
			slog.Error("cannot update agent sync status", "agent", evt.Agent, "err", err.Error())
			return
		}
	})

	return UseCases{
		FindByID: NewFindByID(repo),
		Update:   NewUpdate(&mutex, bus, repo),
	}
}
