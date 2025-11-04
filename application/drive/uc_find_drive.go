// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"fmt"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewFindDrive(repo Repository, globalRootRepo NamedRootRepository, userRootRepo UserRootRepository) FindDrive {
	return func(subject auth.Subject, fid FID) (option.Opt[Drive], error) {
		loopCheck := 0
		parent := fid
		for {
			loopCheck++
			if loopCheck > 1000 {
				return option.Opt[Drive]{}, fmt.Errorf("recursion error detected while traversing parents: check for cycle")
			}

			optRoot, err := repo.FindByID(parent)
			if err != nil {
				return option.Opt[Drive]{}, err
			}

			if optRoot.IsNone() {
				return option.Opt[Drive]{}, nil // we found a stale reference, has it been removed while iterating?
			}

			root := optRoot.Unwrap()
			if root.Parent == "" {
				if !root.CanRead(subject) {
					return option.Opt[Drive]{}, user.PermissionDeniedErr
				}

				// traverse globals which is o(n)
				for namedRoot, err := range globalRootRepo.All() {
					if err != nil {
						return option.Opt[Drive]{}, err
					}

					if namedRoot.Root == root.ID {
						return option.Some(Drive{
							Namespace: NamespaceGlobal,
							Name:      namedRoot.ID,
							Root:      root.ID,
						}), nil
					}
				}

				// traverse users
				for roots, err := range userRootRepo.All() {
					if err != nil {
						return option.Opt[Drive]{}, err
					}

					for id, fid := range roots.Roots {
						if fid == root.ID {
							return option.Some(Drive{
								Namespace: NamespacePrivate,
								Name:      id,
								Root:      fid,
							}), nil
						}
					}
				}

				// not found, probably a stale ref
				return option.Opt[Drive]{}, nil
			}

			parent = optRoot.Unwrap().Parent
		}
	}
}
