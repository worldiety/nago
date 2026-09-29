// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"fmt"

	"go.wdy.de/nago/auth"
)

func NewUpdateSettings(m *Manager, repo SettingsRepository) UpdateSettings {
	return func(subject auth.Subject, settings Settings) error {
		if err := subject.Audit(PermUpdateSettingsByID); err != nil {
			return err
		}

		if !m.Has(settings.ID) {
			return fmt.Errorf("service with id %s not found", settings.ID)
		}

		if err := repo.Save(settings); err != nil {
			return err
		}

		m.Wake(settings.ID)
		return nil
	}
}
