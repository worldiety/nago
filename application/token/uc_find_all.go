// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"iter"
)

func NewFindAll(repo Repository) FindAll {
	return func(subject auth.Subject) iter.Seq2[Token, error] {
		return func(yield func(Token, error) bool) {
			if !subject.Valid() {
				yield(Token{}, user.InvalidSubjectErr)
				return
			}

			for token, err := range repo.All() {
				if err != nil {
					if !yield(Token{}, err) {
						return
					}

					continue
				}

				// either the subject can see all tokens or a specific one, or it is the owner of a token
				allowedToView := subject.HasResourcePermission(repo.Name(), string(token.ID), PermFindAll) || token.Impersonation.UnwrapOr("") == subject.ID()
				if allowedToView {
					if !yield(token, nil) {
						return
					}
				}
			}
		}
	}
}
