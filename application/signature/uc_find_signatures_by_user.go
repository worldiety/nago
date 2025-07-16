// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package signature

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewFindSignaturesByUser(idx *inMemoryIndex) FindSignaturesByUser {
	return func(subject user.Subject, uid user.ID) iter.Seq2[Signature, error] {
		if subject.ID() != uid {
			if err := subject.Audit(PermFindSignaturesByUser); err != nil {
				return xiter.WithError[Signature](err)
			}
		}

		return idx.ByUser(uid)
	}
}
