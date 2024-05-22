package core

import "go.wdy.de/nago/presentation/ora"

// TODO make me an interface?

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

func (n *NavigationController) Reload() {
	// TODO this does not change history stack but destroys the stack entirely
}
