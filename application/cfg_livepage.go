package application

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

func (c *Configurator) Component(id ora.ComponentFactoryId, factory func(wnd core.Window) core.Component) {
	c.uiApp.Components[id] = factory
}
