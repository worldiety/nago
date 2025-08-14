// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"iter"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewFindDeclaredWorkflows(declarations *concurrent.RWMap[ID, *workflow]) FindDeclaredWorkflows {
	return func(subject user.Subject) iter.Seq2[DeclareOptions, error] {
		return func(yield func(DeclareOptions, error) bool) {
			for id, decl := range declarations.All() {
				if !subject.HasResourcePermission(RepositoryNameDeclaredWorkflows, string(id), PermFindDeclaredWorkflows) {
					continue
				}

				if !yield(decl.opts, nil) {
					return
				}
			}
		}
	}
}
