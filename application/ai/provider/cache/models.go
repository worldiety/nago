// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cache

import (
	"iter"

	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Models = (*cacheModels)(nil)

type cacheModels struct {
	parent *Provider
}

func (c cacheModels) All(subject auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) {
		for m, err := range c.parent.repoModels.All() {
			if err != nil {
				if !yield(m, err) {
					return
				}

				continue
			}

			if !subject.HasResourcePermission(c.parent.repoModels.Name(), string(m.ID), PermFindAllModel) {
				continue
			}

			if !yield(m, nil) {
				return
			}
		}
	}
}
