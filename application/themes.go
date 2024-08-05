package application

import (
	"go.wdy.de/nago/presentation/core"
)

func (c *Configurator) ColorSet(scheme core.ColorScheme, cs core.ColorSet) {

	c.colorSets[scheme][cs.Namespace()] = cs
}

/*// CreateThemes takes the given 3 colors and generates all required themes from it.
// Remember the following color definitions:
// * primary is used for the overall color impression. Most areas colors are derived from different brightness values.
// * secondary is used for some accent colors and derived mostly by different transparency levels.
// * tertiary is the color for interactive elements like buttons.
func CreateThemes(primary, secondary, tertiary ui.Color) ora.Themes {
	themes := ora.Themes{
		Dark: ora.Theme{
			Colors: map[ora.NamespaceName]map[string]ora.Color{},
		},
		Light: ora.Theme{
			Colors: map[ora.NamespaceName]map[string]ora.Color{},
		},
	}

	light := ui.DefaultColors(core.Light, primary, secondary, tertiary)
	dark := ui.DefaultColors(core.Dark, primary, secondary, tertiary)

	themes.Light.Colors[ora.NamespaceName(light.Namespace())] = ConvertColorSetToMap(light)
	themes.Dark.Colors[ora.NamespaceName(dark.Namespace())] = ConvertColorSetToMap(dark)

	return themes
}
*/
