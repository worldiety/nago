// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"log/slog"

	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
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

		uc := c.themeManagement.UseCases
		events.SubscribeFor[theme.SettingsUpdated](c.EventBus(), func(theme.SettingsUpdated) {
			// The lock orders this update with building the application, see newCoreApplication. The bus delivers
			// asynchronously, so events may arrive out of order: the stored theme is applied rather than the one of
			// the event, thus the last update always wins.
			c.themeMutex.Lock()
			defer c.themeMutex.Unlock()

			app := c.app.Load()
			if app == nil {
				return // the application reads the stored theme when it is built
			}

			colors, err := uc.ReadColors(user.SU())
			if err != nil {
				slog.Error("cannot read the theme colors", "err", err)
				return
			}

			fonts, err := uc.ReadFonts(user.SU())
			if err != nil {
				slog.Error("cannot read the theme fonts", "err", err)
				return
			}

			app.UpdateColorSet(core.Dark, colors.Dark)
			app.UpdateColorSet(core.Light, colors.Light)
			app.UpdateFonts(fonts)
		})
	}

	return *c.themeManagement, nil
}
