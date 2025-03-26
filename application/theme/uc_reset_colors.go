// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package theme

import (
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
)

func NewResetColors(bus events.Bus, loadGlobal settings.LoadGlobal, storeGlobal settings.StoreGlobal) ResetColors {
	return func(subject auth.Subject) error {
		if err := subject.Audit(PermUpdateColors); err != nil {
			return err
		}

		cfg := settings.ReadGlobal[Settings](loadGlobal)
		cfg.Colors = Colors{}
		err := storeGlobal(user.SU(), cfg)
		if err != nil {
			return err
		}

		return nil
	}
}
