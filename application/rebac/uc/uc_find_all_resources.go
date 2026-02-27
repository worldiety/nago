// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ucrebac

import (
	"iter"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xiter"
)

func NewFindAllResources(rdb *rebac.DB) FindAllResources {
	return func(subject user.Subject) iter.Seq2[rebac.Resources, error] {
		if err := subject.Audit(PermFindAllResources); err != nil {
			return xiter.WithError[rebac.Resources](err)
		}

		return func(yield func(rebac.Resources, error) bool) {
			for resources := range rdb.AllResources() {
				if !yield(resources, nil) {
					return
				}
			}
		}
	}
}
