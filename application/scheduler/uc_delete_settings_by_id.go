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

func NewDeleteSettingsByID(m *Manager, repo SettingsRepository) DeleteSettingsByID {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermDeleteSettingsByID); err != nil {
			return err
		}

		before, _, known := m.effectiveSettings(id)
		if err := repo.DeleteByID(id); err != nil {
			return err
		}

		// the defaults apply immediately
		if known {
			m.settingsChanged(id, before)
		}
		return nil
	}
}
