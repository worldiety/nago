// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"iter"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewReadDrives(globalRootRepo NamedRootRepository, userRootRepo UserRootRepository) ReadDrives {
	return func(subject auth.Subject, uid user.ID) iter.Seq2[Drive, error] {
		return func(yield func(Drive, error) bool) {

			for root, err := range globalRootRepo.All() {
				if err != nil {
					yield(Drive{}, err)
					return
				}

				if root.Root != "" {
					if !yield(Drive{
						Namespace: NamespaceGlobal,
						Name:      root.ID,
						Root:      root.Root,
					}, nil) {
						return
					}
				}
			}

			// security note: this looks quite dangerous. However, users should probably work with groups in the drive
			if subject.ID() == uid || subject.HasPermission(PermOpenFile) {
				optRoot, err := userRootRepo.FindByID(uid)
				if err != nil {
					yield(Drive{}, err)
					return
				}

				if optRoot.IsSome() {
					for name, id := range optRoot.Unwrap().Roots {
						if !yield(Drive{
							Namespace: NamespacePrivate,
							Name:      name,
							Root:      id,
						}, nil) {
							return
						}
					}
				}
			}

			// TODO find shared with user through some inverse lookup
		}

	}
}
