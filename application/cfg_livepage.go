package application

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
	"go.wdy.de/nago/presentation/ui"
)

// deprecated: use Component instead
func (c *Configurator) Page(id ui.PageID, factory func(wire ui.Wire) *ui.Page) {
	c.uiApp.LivePages[id] = factory
	c.uiApp.Components[protocol.ComponentFactoryId(id)] = func(wire ui.Wire) core.Component {
		return factory(wire)
	}
}

func (c *Configurator) Component(id protocol.ComponentFactoryId, factory func(wire ui.Wire) core.Component) {
	c.uiApp.Components[id] = factory
}
