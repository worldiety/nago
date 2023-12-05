package application

import "go.wdy.de/nago/presentation/ui"

func (c *Configurator) Page(id ui.PageID, factory func(wire ui.Wire) *ui.Page) {
	c.uiApp.LivePages[id] = factory
}
