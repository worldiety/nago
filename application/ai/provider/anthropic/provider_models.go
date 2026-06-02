// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package anthropic

import (
	"iter"

	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Models = (*anthropicModels)(nil)

type anthropicModels struct {
	parent *anthropicProvider
}

func (m *anthropicModels) All(subject auth.Subject) iter.Seq2[model.Model, error] {
	return m.parent.listModels(subject)
}

// listModels is the shared implementation used by both provider.Models and completion.Completions.Models.
func (p *anthropicProvider) listModels(subject auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) {
		models, err := p.client().ListModels()
		if err != nil {
			yield(model.Model{}, err)
			return
		}

		for _, m := range models {
			name := m.DisplayName
			if name == "" {
				name = m.ID
			}

			if !yield(model.Model{ID: model.ID(m.ID), Name: name}, nil) {
				return
			}
		}
	}
}

