// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
	"strings"
)

func NewFindByMail(repository Repository) FindByMail {
	return func(subject permission.Auditable, email Email) (std.Option[User], error) {
		if err := subject.Audit(PermFindByMail); err != nil {
			return std.None[User](), err
		}

		// normalize given mail
		email = Email(strings.TrimSpace(strings.ToLower(string(email))))

		// do not introduce the global mutex here, because they are not reentrant you likely get a deadlock
		// TODO this is really slow O(n), we either need some cache or an inverse index
		var consistencyCheck []User
		for user, err := range repository.All() {
			if err != nil {
				return std.None[User](), fmt.Errorf("cannot loop user repo: %w", err)
			}

			if user.Email == email {
				consistencyCheck = append(consistencyCheck, user)
			}
		}

		if len(consistencyCheck) == 0 {
			return std.None[User](), nil
		}

		if len(consistencyCheck) == 1 {
			return std.Some(consistencyCheck[0]), nil
		}

		return std.None[User](), fmt.Errorf("unique mail violation: multiple users for email %v", email)
	}
}
