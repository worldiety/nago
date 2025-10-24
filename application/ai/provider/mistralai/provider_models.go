// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"iter"

	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Models = (*mistralModels)(nil)

type mistralModels struct {
	parent *mistralProvider
}

func (p *mistralModels) client() *Client {
	return p.parent.client()
}

func (p *mistralModels) All(subject auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) {
		models, err := p.client().ListAllModels(subject.Language())
		if err != nil {
			yield(model.Model{}, err)
			return
		}

		for _, m := range models {
			if !yield(m.IntoModel(), nil) {
				return
			}
		}
	}
}
