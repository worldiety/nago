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
func CreateThemes(primary, secondary, tertiary ui.Color) proto.Themes {
	themes := proto.Themes{
		Dark: proto.Theme{
			Colors: map[proto.NamespaceName]map[string]proto.Color{},
		},
		Light: proto.Theme{
			Colors: map[proto.NamespaceName]map[string]proto.Color{},
		},
	}

	light := ui.DefaultColors(core.Light, primary, secondary, tertiary)
	dark := ui.DefaultColors(core.Dark, primary, secondary, tertiary)

	themes.Light.Colors[proto.NamespaceName(light.Namespace())] = ConvertColorSetToMap(light)
	themes.Dark.Colors[proto.NamespaceName(dark.Namespace())] = ConvertColorSetToMap(dark)

	return themes
}
*/
