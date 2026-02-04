// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"os"
	"slices"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewFindFormShareByID(idx *shareIndex) FindFormShareByID {
	return func(subject auth.Subject, shareID FormShareID) (option.Opt[FormShare], error) {
		var optShare option.Opt[FormShare]
		err := idx.mut(func() error {
			share, ok := idx.lookup[shareID]
			if !ok {
				return fmt.Errorf("share %s does not exist: %w", shareID, os.ErrNotExist)
			}

			optWs, err := idx.load(user.SU(), share.Workspace)
			if err != nil {
				return fmt.Errorf("failed to load workspace: %w", err)
			}

			if optWs.IsNone() {
				return nil
			}

			ws := optWs.Unwrap()
			if subject.HasPermission(PermFindWorkspaces) || ws.IsOwner(subject.ID()) || share.AllowUnauthenticated {
				optShare = option.Some(share.clone())
			}

			if subject.Valid() && !share.AllowUnauthenticated && len(share.AllowedUsers) == 0 {
				optShare = option.Some(share.clone())
			}

			if slices.Contains(share.AllowedUsers, subject.ID()) {
				optShare = option.Some(share.clone())
			}

			return nil
		})

		return optShare, err
	}
}
