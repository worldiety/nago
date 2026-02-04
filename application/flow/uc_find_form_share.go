// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"errors"
	"os"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewFindFormShare(idx *shareIndex) FindFormShare {
	return func(subject auth.Subject, workspace WorkspaceID, form FormID) (option.Opt[FormShare], error) {
		var wrk *Workspace
		err := idx.byForm(workspace, form, func(ws *Workspace, form *Form) error {
			wrk = ws
			return nil
		})

		if errors.Is(err, os.ErrNotExist) {
			return option.None[FormShare](), nil
		}

		if err != nil {
			return option.None[FormShare](), err
		}

		share, ok := idx.reverseLookup[form]
		if !ok {
			return option.None[FormShare](), nil
		}

		isAllowed := subject.HasPermission(PermFindWorkspaces) || wrk.IsOwner(subject.ID()) || share.AllowUnauthenticated
		if !isAllowed {
			return option.None[FormShare](), user.PermissionDeniedError("workspace or share forbids access")
		}

		return option.Some(share.clone()), nil
	}
}
