// Copyright (c) 2026 worldiety GmbH
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
	"go.wdy.de/nago/presentation/core"
)

func NewUpdateFonts(bus events.Bus, loadGlobal settings.LoadGlobal, storeGlobal settings.StoreGlobal) UpdateFonts {
	return func(subject auth.Subject, fonts core.Fonts) error {
		// Used colors permission as it is done the same way in ReadFonts. TODO: Change this?
		if err := subject.Audit(PermUpdateColors); err != nil {
			return err
		}

		cfg := settings.ReadGlobal[Settings](loadGlobal)
		cfg.Fonts = fonts
		err := storeGlobal(user.SU(), cfg)
		if err != nil {
			return err
		}

		bus.Publish(SettingsUpdated{
			Settings: cfg,
		})

		return nil
	}
}
