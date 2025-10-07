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

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

func NewWalkDir(repo Repository) WalkDir {
	return func(subject auth.Subject, root FID, walker func(fid FID, file File, err error) error) error {
		return walkDir(repo, subject, root, walker)
	}
}

func walkDir(repo Repository, subject auth.Subject, root FID, walker func(fid FID, file File, err error) error) error {
	var zero File
	optFile, err := readFileStat(repo, root)
	if err != nil {
		if err := walker(root, zero, err); err != nil {
			return err
		}

		return nil // the walker just skipped the error, but we cannot continue
	}

	if optFile.IsNone() {
		if err := walker(root, zero, os.ErrNotExist); err != nil {
			return err
		}

		return nil // the walker just skipped, but we can't continue
	}

	file := optFile.Unwrap()
	if !file.CanRead(subject) {
		if err := walker(root, zero, fmt.Errorf("permission denied: %w: %w)", user.PermissionDeniedErr, os.ErrPermission)); err != nil {
			return err
		}

		return nil // the walker skipped, but we are not allowed to continue
	}

	// walk the root regular
	if err := walker(root, file, nil); err != nil {
		return err
	}

	// walk recursive
	for fid := range file.Entries.All() {
		if err := walkDir(repo, subject, fid, walker); err != nil {
			return err
		}
	}

	return nil
}
