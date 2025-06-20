// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"go.wdy.de/nago/application/dataimport/parser"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
	"slices"
	"strings"
)

func NewFindParsers(m *concurrent.RWMap[parser.ID, parser.Parser]) FindParsers {
	return func(subject auth.Subject) iter.Seq2[parser.Parser, error] {
		if err := subject.Audit(PermFindImporters); err != nil {
			return xiter.WithError[parser.Parser](err)
		}

		return func(yield func(parser.Parser, error) bool) {
			var tmp []parser.Parser
			for _, imp := range m.All() {
				tmp = append(tmp, imp)
			}

			slices.SortFunc(tmp, func(a, b parser.Parser) int {
				return strings.Compare(a.Configuration().Name, b.Configuration().Name)
			})

			for _, p := range tmp {
				if !yield(p, nil) {
					return
				}
			}
		}
	}
}
