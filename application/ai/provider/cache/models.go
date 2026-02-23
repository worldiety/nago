// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cache

import (
	"context"
	"iter"
	"log/slog"
	"slices"
	"strings"

	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

var _ provider.Models = (*cacheModels)(nil)

type cacheModels struct {
	parent *Provider
}

func (c cacheModels) All(subject auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) {
		ctx := context.Background()
		num, err := c.parent.idxProvModels.CountByPrimary(ctx, c.parent.Identity()) // this must never be empty
		if err != nil {
			yield(model.Model{}, err)
			return
		}

		if num == 0 {
			for mod, err := range c.parent.prov.Models().All(subject) {
				if err != nil {
					yield(model.Model{}, err)
					return
				}

				if mod.ID == "" {
					slog.Warn("provider returned ai model without an ID")
					continue
				}

				if err := c.parent.repoModels.Save(mod); err != nil {
					yield(model.Model{}, err)
					return
				}

				if err := c.parent.idxProvModels.Put(c.parent.Identity(), mod.Identity()); err != nil {
					yield(model.Model{}, err)
					return
				}
			}
		}

		var tmp []model.Model
		for key, err := range c.parent.idxProvModels.AllByPrimary(ctx, c.parent.Identity()) {
			if err != nil {
				yield(model.Model{}, err)
				return
			}

			optMod, err := c.parent.repoModels.FindByID(key.Secondary)
			if err != nil {
				yield(model.Model{}, err)
				return
			}

			if optMod.IsNone() {
				continue
			}

			m := optMod.Unwrap()

			if !subject.HasResourcePermission(rebac.Namespace(c.parent.repoModels.Name()), rebac.Instance(m.ID), PermFindAllModel) {
				continue
			}

			tmp = append(tmp, m)
		}

		slices.SortFunc(tmp, func(a, b model.Model) int {
			return strings.Compare(a.Name, b.Name)
		})

		for _, m := range tmp {
			if !yield(m, nil) {
				return
			}
		}
	}
}
