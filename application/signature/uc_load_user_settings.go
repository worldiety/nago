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

func NewLoadUserSettings(repo UserSettingsRepository) LoadUserSettings {
	return func(subject user.Subject, uid user.ID) (UserSettings, error) {
		if uid != subject.ID() {
			return UserSettings{}, fmt.Errorf("only a user itself can load its user signature settings: %w", user.PermissionDeniedErr)
		}

		optS, err := repo.FindByID(uid)
		if err != nil {
			return UserSettings{}, err
		}

		if optS.IsNone() {
			return UserSettings{}, nil
		}

		return optS.Unwrap(), nil
	}
}
