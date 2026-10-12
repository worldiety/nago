// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package theme

import (
	"fmt"

	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
)

func NewReadCustomization(loadGlobal settings.LoadGlobal, defaults Defaults) ReadCustomization {
	return func(subject auth.Subject) (Customization, error) {
		if err := subject.Audit(PermReadColors); err != nil {
			return Customization{}, err
		}

		cfg := settings.ReadGlobal[Settings](loadGlobal)
		if cfg.Customization.Valid() {
			return cfg.Customization, nil
		}

		return CustomizationOf(withDefaults(cfg.Colors, defaults)), nil
	}
}

func NewUpdateCustomization(bus events.Bus, loadGlobal settings.LoadGlobal, storeGlobal settings.StoreGlobal) UpdateCustomization {
	return func(subject auth.Subject, customization Customization) error {
		if err := subject.Audit(PermUpdateColors); err != nil {
			return err
		}

		if err := customization.Light.validate(); err != nil {
			return fmt.Errorf("light mode: %w", err)
		}

		if err := customization.Dark.validate(); err != nil {
			return fmt.Errorf("dark mode: %w", err)
		}

		cfg := settings.ReadGlobal[Settings](loadGlobal)
		cfg.Customization = customization
		cfg.Colors = customization.Colors()
		if err := storeGlobal(user.SU(), cfg); err != nil {
			return err
		}

		bus.Publish(SettingsUpdated{
			Settings: cfg,
		})

		return nil
	}
}
