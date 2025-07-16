// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package signature

import (
	"fmt"
	"go.wdy.de/nago/application/user"
)

func NewUpdateUserSettings(repo UserSettingsRepository) UpdateUserSettings {
	return func(subject user.Subject, uid user.ID, cdata UpdateUserSettingsData) error {
		if uid != subject.ID() {
			return fmt.Errorf("only a user itself can updates its user signature settings: %w", user.PermissionDeniedErr)
		}

		return repo.Save(UserSettings{
			User:           uid,
			ImageSignature: cdata.ImageSignature,
		})
	}
}
