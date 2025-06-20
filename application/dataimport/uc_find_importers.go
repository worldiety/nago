// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewFindImporters(m *concurrent.RWMap[importer.ID, importer.Importer]) FindImporters {
	return func(subject auth.Subject) iter.Seq2[importer.Importer, error] {
		if err := subject.Audit(PermFindImporters); err != nil {
			return xiter.WithError[importer.Importer](err)
		}

		return func(yield func(importer.Importer, error) bool) {
			for _, imp := range m.All() {
				if !yield(imp, nil) {
					return
				}
			}
		}
	}
}
