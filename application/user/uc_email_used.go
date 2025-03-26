// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

func NewEMailUsed(repo Repository) EMailUsed {
	return func(email Email) (bool, error) {
		for user, err := range repo.All() {
			if err != nil {
				return false, err
			}

			if user.Email.Equals(email) {
				return true, nil
			}
		}

		return false, nil
	}
}
