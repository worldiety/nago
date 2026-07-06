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

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewReadFileGrants(repo Repository, rdb *rebac.DB) ReadFileGrants {
	return func(subject auth.Subject, fid FID) ([]FileGrant, error) {
		optFile, err := readFileStat(repo, fid)
		if err != nil {
			return nil, fmt.Errorf("cannot read file %s: %w", fid, err)
		}

		if optFile.IsNone() {
			return nil, fmt.Errorf("file does not exist: %s: %w", fid, os.ErrNotExist)
		}

		file := optFile.Unwrap()
		if !file.CanRead(subject) {
			return nil, fmt.Errorf("cannot read file grants %s: %w", fid, user.PermissionDeniedErr)
		}

		return readFileGrants(rdb, fid)
	}
}
