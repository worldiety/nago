// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
)

func NewFindSettingsByID(repo SettingsRepository) FindSettingsByID {
	return func(subject auth.Subject, id ID) (std.Option[Settings], error) {
		if err := subject.Audit(PermFindSettingsByID); err != nil {
			return std.None[Settings](), err
		}

		return repo.FindByID(id)
	}
}
