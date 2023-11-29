package application

import "go.wdy.de/nago/presentation/ui"

func (c *Configurator) LivePage(id ui.PageID, factory func(wire ui.Wire) *ui.LivePage) {
	c.uiApp.LivePages[id] = factory
}
