// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"iter"
	"log/slog"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/pkg/xiter"
)

func NewListResourcePermissions(rdb *rebac.DB) ListResourcePermissions {
	return func(subject AuditableUser, uid ID) iter.Seq2[ResourceWithPermissions, error] {
		if err := subject.Audit(PermListResourcePermissions); err != nil {
			return xiter.WithError[ResourceWithPermissions](err)
		}

		return func(yield func(ResourceWithPermissions, error) bool) {
			// note that this is semantically not the same as before, I don't know how implement this in a meaningful manner
			slog.Error("ListResourcePermissions is not implemented anymore, migrate to rebac")
		}
	}
}
