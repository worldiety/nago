package application

import (
	"go.wdy.de/nago/presentation/ora"
)

func (c *Configurator) ColorSet(scheme ora.ColorScheme, cs ora.ColorSet) {

	c.colorSets[scheme][cs.Namespace()] = cs
}
