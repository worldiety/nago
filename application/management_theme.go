// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
)

// ThemeManagement is a nago system(Theme Management).
// Theme Management handles the configuration of theme and corporate identity settings.
//
// It allows to define logos and app icons (for dark and light mode), configure
// legal information (e.g. Impressum, Privacy Policy, Terms, User Agreement),
// and set provider contact details such as responsible entity, contact email,
// and API documentation URL.
//
// Additionally, developers can define fonts and base colors (main, interactive, accent)
// directly via code. Colors can differ between dark and light mode.
type ThemeManagement struct {
	UseCases theme.UseCases
}

func (c *Configurator) ThemeManagement() (ThemeManagement, error) {
	if c.themeManagement == nil {
		sets, err := c.SettingsManagement()
		if err != nil {
			return ThemeManagement{}, err
		}
		c.themeManagement = &ThemeManagement{UseCases: theme.NewUseCases(
			c.EventBus(),
			sets.UseCases.LoadGlobal,
			sets.UseCases.StoreGlobal,
		)}

		events.SubscribeFor[theme.SettingsUpdated](c.EventBus(), func(evt theme.SettingsUpdated) {
			// TODO this entire thing is racy if updated after Configurator.Run
			if c.app != nil {
				c.app.UpdateColorSet(core.Dark, evt.Settings.Colors.Dark)
				c.app.UpdateColorSet(core.Light, evt.Settings.Colors.Light)
			}
		})
	}

	return *c.themeManagement, nil
}
