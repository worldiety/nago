package core

import "go.wdy.de/nago/presentation/ora"

// TODO make me an interface?

type NavigationController struct {
	destroyed bool
	scope     *Scope
}

func NewNavigationController(scope *Scope) *NavigationController {
	return &NavigationController{
		scope: scope,
	}
}

func (n *NavigationController) ForwardTo(id NavigationPath, values Values) {
	if n.destroyed {
		return
	}

	n.scope.Publish(ora.NavigationForwardToRequested{
		Type:    ora.NavigationForwardToRequestedT,
		Factory: ora.ComponentFactoryId(id),
		Values:  values,
	})
}

func (n *NavigationController) Back() {
	if n.destroyed {
		return
	}

	n.scope.Publish(ora.NavigationBackRequested{
		Type: ora.NavigationBackRequestedT,
	})
}

func (n *NavigationController) ResetTo(id NavigationPath, values Values) {
	if n.destroyed {
		return
	}

	n.scope.Publish(ora.NavigationResetRequested{
		Type:    ora.NavigationResetRequestedT,
		Factory: ora.ComponentFactoryId(id),
		Values:  values,
	})
}

func (n *NavigationController) Reload() {
	if n.destroyed {
		return
	}

	n.scope.Publish(ora.NavigationReloadRequested{
		Type: ora.NavigationReloadRequestedT,
	})
}
