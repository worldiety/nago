// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workspace

import (
	"iter"

	"go.wdy.de/nago/auth"
)

func NewFindAll(repo Repository) FindAll {
	return func(subject auth.Subject) iter.Seq2[ID, error] {
		return func(yield func(ID, error) bool) {
			for ws, err := range repo.All() {
				if err != nil {
					if !yield("", err) {
						return
					}

					continue
				}

				if !subject.HasResourcePermission(repo.Name(), string(ws.ID), PermFindAll) {
					continue
				}

				if !yield(ws.ID, nil) {
					return
				}
			}
		}
	}
}
