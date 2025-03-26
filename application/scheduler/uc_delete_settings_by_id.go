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

func NewDeleteSettingsByID(repo SettingsRepository) DeleteSettingsByID {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermDeleteSettingsByID); err != nil {
			return err
		}

		return repo.DeleteByID(id)
	}
}
