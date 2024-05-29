package application

import "go.wdy.de/nago/presentation/ora"

func (c *Configurator) SetThemes(themes ora.Themes) {
	c.themes = themes
}

func (c *Configurator) SetLightTheme(theme ora.Theme) {
	c.themes.Light = theme
}

func (c *Configurator) SetDarkTheme(theme ora.Theme) {
	c.themes.Dark = theme
}
