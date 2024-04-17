package core

import "go.wdy.de/nago/presentation/ora"

type Values map[string]string

type NavigationController struct {
	scope *Scope
}

func NewNavigationController(scope *Scope) *NavigationController {
	return &NavigationController{
		scope: scope,
	}
}

func (n *NavigationController) ForwardTo(id ora.ComponentFactoryId, values Values) {
	n.scope.Publish(ora.NavigationForwardToRequested{
		Type:    ora.NavigationForwardToRequestedT,
		Factory: id,
		Values:  values,
	})
}

func (n *NavigationController) Back() {
	n.scope.Publish(ora.NavigationBackRequested{
		Type: ora.NavigationBackRequestedT,
	})
}

func (n *NavigationController) ResetTo(id ora.ComponentFactoryId, values Values) {
	n.scope.Publish(ora.NavigationResetRequested{
		Type:    ora.NavigationResetRequestedT,
		Factory: id,
		Values:  values,
	})
}
