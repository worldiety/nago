// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package signature

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
)

func NewFindByID(repo Repository, idx *inMemoryIndex) FindByID {
	return func(subject user.Subject, id ID) (option.Opt[Signature], error) {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(id), PermFindByID); err != nil {
			return option.Opt[Signature]{}, err
		}

		sig, ok := idx.ByID(id)
		if ok {
			return option.Some(sig), nil
		}

		return option.Opt[Signature]{}, nil
	}
}
