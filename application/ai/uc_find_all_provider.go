// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ai

import (
	"iter"
	"slices"
	"strings"

	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewFindAllProvider(m *concurrent.RWMap[provider.ID, provider.Provider]) FindAllProvider {
	return func(subject auth.Subject) iter.Seq2[provider.Provider, error] {
		return func(yield func(provider.Provider, error) bool) {
			if err := subject.Audit(PermFindAllProvider); err != nil {
				yield(nil, err)
				return
			}

			var tmp []provider.Provider // keep stable order
			for _, p := range m.All() {
				tmp = append(tmp, p)
			}

			slices.SortFunc(tmp, func(a, b provider.Provider) int {
				return strings.Compare(a.Name(), b.Name())
			})

			for _, p := range tmp {
				if !yield(p, nil) {
					return
				}
			}

		}
	}
}
