// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
)

func NewFindByMail(repository Repository, idx *UserIndex) FindByMail {
	return func(subject permission.Auditable, email Email) (std.Option[User], error) {
		if err := subject.Audit(PermFindByMail); err != nil {
			return std.None[User](), err
		}

		// do not introduce the global mutex here, because they are not reentrant you likely get a deadlock
		id, ok, err := idx.LookupMail(email)
		if err != nil {
			return std.None[User](), fmt.Errorf("cannot lookup mail index: %w", err)
		}

		if !ok {
			return std.None[User](), nil
		}

		return repository.FindByID(id)
	}
}
