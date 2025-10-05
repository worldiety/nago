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

func NewStat(repo Repository) Stat {
	return func(subject auth.Subject, fid FID) (option.Opt[File], error) {
		optFile, err := readFileStat(repo, fid)
		if err != nil || optFile.IsNone() {
			return optFile, err
		}

		file := optFile.Unwrap()

		if file.CanRead(subject) {
			return optFile, nil
		}

		return option.None[File](), fmt.Errorf("not allowed to read file: %s: %w", fid, user.PermissionDeniedErr)
	}
}

func readFileStat(repo Repository, fid FID) (option.Opt[File], error) {
	optFile, err := repo.FindByID(fid)
	if err != nil {
		return option.None[File](), fmt.Errorf("failed to load from repo: %w", err)
	}

	if optFile.IsNone() {
		return option.None[File](), nil
	}

	file := optFile.Unwrap()
	file.repo = repo

	return option.Some(file), nil
}
