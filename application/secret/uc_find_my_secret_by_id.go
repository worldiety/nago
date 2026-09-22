// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
)

func NewFindMySecretByID(repository Repository) FindMySecretByID {
	return func(subject auth.Subject, id ID) (std.Option[Secret], error) {
		if err := subject.Audit(PermFindMySecrets); err != nil {
			return std.Option[Secret]{}, err
		}

		optSecret, err := repository.FindByID(id)
		if err != nil {
			return std.Option[Secret]{}, err
		}

		if optSecret.IsNone() {
			return optSecret, nil
		}

		if !optSecret.Unwrap().HasAccess(subject) {
			return std.Option[Secret]{}, AccessDeniedErr
		}

		return optSecret, nil
	}
}
