package application

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui"
)

// deprecated: use Component instead
func (c *Configurator) Page(id ui.PageID, factory func(wire ui.Wire) *ui.Page) {
	c.uiApp.LivePages[id] = factory
	c.uiApp.Components[ora.ComponentFactoryId(id)] = func(realm core.Window) core.Component {
		return factory(noOpWireStub{})
	}
}

func (c *Configurator) Component(id ora.ComponentFactoryId, factory func(realm core.Window) core.Component) {
	c.uiApp.Components[id] = factory
}
