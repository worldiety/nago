// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Ident: Custom-License

package environment

import (
	"iter"
	"slices"

	"go.wdy.de/nago/auth"
)

func NewFindAll(repo Repository) FindAll {
	return func(subject auth.Subject) iter.Seq2[Environment, error] {
		return func(yield func(Environment, error) bool) {
			for env, err := range repo.All() {
				if err != nil {
					if !yield(Environment{}, err) {
						return
					}

					continue
				}

				if slices.Contains(env.Owner, subject.ID()) {
					if !yield(env, nil) {
						return
					}
				}
			}
		}
	}
}
