// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"log/slog"
	"os"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewResolveTokenRights(repo Repository, findGroupByID group.FindByID, findRoleByID role.FindByID, findUserByID user.FindByID) ResolveTokenRights {
	return func(subject auth.Subject, id ID) (ResolvedTokenRights, error) {
		if !subject.Valid() {
			return ResolvedTokenRights{}, user.InvalidSubjectErr
		}

		optToken, err := repo.FindByID(id)
		if err != nil {
			return ResolvedTokenRights{}, err
		}

		if optToken.IsNone() {
			return ResolvedTokenRights{}, os.ErrNotExist
		}

		token := optToken.Unwrap()
		allowedToView := subject.HasResourcePermission(rebac.Namespace(repo.Name()), rebac.Instance(token.ID), PermResolveTokenRights) || token.Impersonation.UnwrapOr("") == subject.ID()
		if !allowedToView {
			return ResolvedTokenRights{}, user.PermissionDeniedErr
		}

		var res ResolvedTokenRights

		// security note: always keep either or implementation for impersonation
		slog.Error("ResolveTokenRights is not implemented anymore, use rebac api")

		return res, nil
	}
}
