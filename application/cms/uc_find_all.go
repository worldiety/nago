// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"iter"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

func NewFindAll(repo Repository) FindAll {
	return func(subject auth.Subject) iter.Seq2[*Document, error] {
		return func(yield func(*Document, error) bool) {
			for pDoc, err := range repo.All() {
				if err != nil {
					if !yield(nil, err) {
						return
					}

					continue
				}

				if !subject.HasResourcePermission(rebac.Namespace(repo.Name()), rebac.Instance(pDoc.ID), PermFindAll) {
					continue
				}

				doc := pDoc.IntoModel()
				if !yield(doc, nil) {
					return
				}
			}
		}
	}
}
