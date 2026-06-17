// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"iter"

	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Models = (*gollamaModels)(nil)

type gollamaModels struct {
	parent *gollamaProvider
}

func (m *gollamaModels) All(subject auth.Subject) iter.Seq2[model.Model, error] {
	return listModels()
}

// listModels yields the curated, hardcoded catalog. It is the shared implementation used by both
// provider.Models and completion.Completions.Models.
func listModels() iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) {
		for _, e := range catalog {
			if !yield(e.toModel(), nil) {
				return
			}
		}
	}
}
