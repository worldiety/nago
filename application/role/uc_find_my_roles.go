// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package role

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/pkg/xslices"
	"iter"
)

func NewFindMyRoles(repository Repository) FindMyRoles {
	return func(subject permission.Auditable) iter.Seq2[Role, error] {
		type roleOwner interface {
			HasRole(ID) bool
			Roles() iter.Seq[ID]
		}

		owner, ok := subject.(roleOwner)
		if !ok {
			return xiter.WithError[Role](fmt.Errorf("subject %T is not Role owner", subject))
		}

		var tmp []Role
		for id := range owner.Roles() {
			optRole, err := repository.FindByID(id)
			if err != nil {
				return xiter.WithError[Role](err)
			}

			if optRole.IsSome() {
				tmp = append(tmp, optRole.Unwrap())
			}
		}

		return xslices.Values2[[]Role, Role, error](tmp)
	}
}
