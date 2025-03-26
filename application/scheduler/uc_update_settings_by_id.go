// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"go.wdy.de/nago/auth"
)

func NewUpdateSettings(repo SettingsRepository) UpdateSettings {
	return func(subject auth.Subject, settings Settings) error {
		if err := subject.Audit(PermUpdateSettingsByID); err != nil {
			return err
		}

		return repo.Save(settings)
	}
}
