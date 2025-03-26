// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

func NewCountUsers(repo Repository) CountUsers {
	return func() (int, error) {
		n, err := repo.Count()
		if err != nil {
			return 0, err
		}

		return n, nil
	}
}
