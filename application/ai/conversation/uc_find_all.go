// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package conversation

import (
	"iter"

	"go.wdy.de/nago/auth"
)

func NewFindAll(repo Repository) FindAll {
	return func(subject auth.Subject) iter.Seq2[Conversation, error] {
		return func(yield func(Conversation, error) bool) {
			for conversation, err := range repo.All() {
				if err != nil {
					if !yield(conversation, err) {
						return
					}

					continue
				}

				if subject.ID() != conversation.CreatedBy && !subject.HasResourcePermission(repo.Name(), string(conversation.ID), PermFindAll) {
					continue
				}

				if !yield(conversation, nil) {
					return
				}
			}
		}
	}
}
