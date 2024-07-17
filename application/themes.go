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

func (c *Configurator) AddColor(name string, light ora.Color, dark ora.Color) {
	if c.themes.Light.Colors.CustomColors == nil {
		c.themes.Light.Colors.CustomColors = map[string]ora.Color{}
	}

	c.themes.Light.Colors.CustomColors[name] = light

	if c.themes.Dark.Colors.CustomColors == nil {
		c.themes.Dark.Colors.CustomColors = map[string]ora.Color{}
	}

	c.themes.Dark.Colors.CustomColors[name] = dark
}

func (c *Configurator) AddLength(name string, len ora.Length) {
	if c.themes.Light.Lengths.CustomLengths == nil {
		c.themes.Light.Lengths.CustomLengths = map[string]ora.Length{}
	}

	c.themes.Light.Lengths.CustomLengths[name] = len

	if c.themes.Dark.Lengths.CustomLengths == nil {
		c.themes.Dark.Lengths.CustomLengths = map[string]ora.Length{}
	}

	c.themes.Dark.Lengths.CustomLengths[name] = len
}
