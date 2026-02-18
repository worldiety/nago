// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

func NewFindByID(repo Repository) FindByID {
	return func(subject auth.Subject, id ID) (option.Opt[Token], error) {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(id), PermFindAll); err != nil {
			return option.Opt[Token]{}, err
		}

		return repo.FindByID(id)
	}
}
