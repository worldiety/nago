package application

import (
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
)

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
