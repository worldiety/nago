// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewReadDrives(globalRootRepo NamedRootRepository, userRootRepo UserRootRepository) ReadDrives {
	return func(subject auth.Subject, uid user.ID) (Drives, error) {
		drives := Drives{
			Private: map[string]FID{},
			Global:  map[string]FID{},
		}

		for root, err := range globalRootRepo.All() {
			if err != nil {
				return Drives{}, err
			}

			if root.Root != "" {
				drives.Global[root.ID] = root.Root
			}
		}

		if subject.ID() == uid || subject.HasPermission(PermOpenFile) {
			optRoot, err := userRootRepo.FindByID(uid)
			if err != nil {
				return Drives{}, err
			}

			if optRoot.IsSome() {
				for name, id := range optRoot.Unwrap().Roots {
					drives.Private[name] = id
				}
			}
		}

		// TODO find shared with user through some inverse lookup

		return drives, nil
	}
}
