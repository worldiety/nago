// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package group

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
)

func NewFindByID(repo Repository) FindByID {
	return func(subject permission.Auditable, id ID) (std.Option[Group], error) {
		if err := subject.Audit(PermFindByID); err != nil {
			return std.None[Group](), err
		}

		return repo.FindByID(id)
	}
}
