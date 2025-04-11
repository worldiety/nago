// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
)

func NewFindByID(repository Repository) FindByID {
	return func(subject permission.Auditable, id ID) (std.Option[User], error) {
		self := false
		if subj, ok := subject.(Subject); ok {
			// TODO cannot remember why the use case requires the permission.Auditable
			self = subj.ID() == id
		}

		if !self {
			if err := subject.Audit(PermFindByID); err != nil {
				return std.None[User](), err
			}
		}

		return repository.FindByID(id)
	}
}
