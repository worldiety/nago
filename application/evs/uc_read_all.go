// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"context"
	"iter"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
)

func NewReadAll[Evt any](perms Permissions, eventStore blob.Store) ReadAll[Evt] {
	return func(subject auth.Subject) iter.Seq2[SeqKey, error] {
		return func(yield func(SeqKey, error) bool) {
			if err := subject.Audit(perms.ReadAll); err != nil {
				yield("", err)
				return
			}

			for key, err := range eventStore.List(context.Background(), blob.ListOptions{Reverse: true}) {
				if err != nil {
					if !yield("", err) {
						return
					}

					continue
				}

				if !yield(SeqKey(key), nil) {
					return
				}
			}
		}
	}
}
