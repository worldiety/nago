// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package signature

import (
	"go.wdy.de/nago/application/user"
	"iter"
)

func NewFindSignaturesByResource(idx *inMemoryIndex) FindSignaturesByResource {
	return func(subject user.Subject, res user.Resource) iter.Seq2[Signature, error] {
		return func(yield func(Signature, error) bool) {
			for sig, err := range idx.ByResource(res) {
				if err != nil {
					panic("unreachable")
				}

				if subject.Valid() && sig.User == subject.ID() {
					if !yield(sig, nil) {
						return
					}

					continue
				}

				if err := subject.AuditResource(res.Name, res.ID, PermFindSignaturesByResource); err == nil {
					if !yield(sig, nil) {
						return
					}
				}
			}
		}
	}
}
