// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package drive

import (
	"iter"
	"slices"

	"go.wdy.de/nago/pkg/xstrings"
)

func applyStandardEntryOrder(repo Repository, seq iter.Seq[FID]) ([]FID, error) {
	var tmp []File
	for fid := range seq {
		optFile, err := repo.FindByID(fid)
		if err != nil {
			return nil, err
		}

		if optFile.IsNone() {
			// stale file, but we must not purge the file
			tmp = append(tmp, File{ID: fid})
		}

		tmp = append(tmp, optFile.Unwrap())
	}

	slices.SortFunc(tmp, func(a, b File) int {
		if a.IsDir() && !b.IsDir() {
			return -1
		}

		if b.IsDir() && !a.IsDir() {
			return 1
		}

		return xstrings.CompareFold(a.Name(), b.Name())
	})

	tmp2 := make([]FID, 0, len(tmp))
	for _, file := range tmp {
		tmp2 = append(tmp2, file.ID)
	}

	return tmp2, nil
}
