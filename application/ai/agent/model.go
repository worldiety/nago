// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package agent

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

// Model represents quality categories of agent models to use. Currently, three classifications are defined
// and an actual implementation must be able to match these. However, to peek through the abstraction and pick
// a specific model, implementations may also support specific model unportable identifiers.
type Model string

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
	System bool `json:"system,omitempty"`
}

func (e Agent) Identity() ID {
	return e.ID
}

type Repository data.Repository[Agent, ID]

type CreateOptions struct {
	Name         string
	Description  string
	Prompt       string
	Model        Model
	Libraries    []library.ID
	Capabilities []Capability
	Temperature  Temperature
	System       bool
}

type Create func(subject auth.Subject, options CreateOptions) (ID, error)
type DeleteByID func(subject auth.Subject, aid ID) error

type FindByID func(subject auth.Subject, id ID) (option.Opt[Agent], error)

type UpdatePrompt func(subject auth.Subject, options CreateOptions) (ID, error)

// ProvideAgent declares a system agent using a pre-defined non-empty identifier.
type ProvideAgent func(agent Agent) error

type UseCases struct {
	Create       Create
	DeleteByID   DeleteByID
	FindByID     FindByID
	UpdatePrompt UpdatePrompt
	ProvideAgent ProvideAgent
}
