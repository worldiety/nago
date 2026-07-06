// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"fmt"
	"os"
	"sync"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewRevokeFileAccess(mutex *sync.Mutex, repo Repository, rdb *rebac.DB) RevokeFileAccess {
	return func(subject auth.Subject, fid FID, grantee Grantee, perms ...permission.ID) error {
		if !grantee.Valid() {
			return fmt.Errorf("grantee must reference exactly one user or group: %w", os.ErrInvalid)
		}

		src, ok := granteeEntity(grantee)
		if !ok {
			return fmt.Errorf("invalid grantee: %w", os.ErrInvalid)
		}

		if len(perms) == 0 {
			return nil
		}

		mutex.Lock()
		defer mutex.Unlock()

		optFile, err := readFileStat(repo, fid)
		if err != nil {
			return fmt.Errorf("cannot read file %s: %w", fid, err)
		}

		if optFile.IsNone() {
			return fmt.Errorf("file does not exist: %s: %w", fid, os.ErrNotExist)
		}

		file := optFile.Unwrap()
		if !mayChangeACL(subject, file) {
			return fmt.Errorf("cannot change acl of file %s: %w", fid, user.PermissionDeniedErr)
		}

		if err := revokeFilePermissions(rdb, src, fid, perms...); err != nil {
			return fmt.Errorf("cannot revoke file access %s: %w", fid, err)
		}

		return nil
	}
}
