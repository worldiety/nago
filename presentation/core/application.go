package core

import (
	"go.wdy.de/nago/presentation/ora"
	"sync"
	"time"
)

type Application struct {
	mutex         sync.Mutex
	scopes        map[ora.ScopeID]*Scope
	factories     map[ora.ComponentFactoryId]ComponentFactory
	scopeLifetime time.Duration
}

func NewApplication(factories map[ora.ComponentFactoryId]ComponentFactory) *Application {
	return &Application{
		scopeLifetime: time.Minute,
		factories:     factories,
		scopes:        map[ora.ScopeID]*Scope{},
	}
}

// Connect either connects an existing scope with the channel or creates a new scope with the given id.
func (a *Application) Connect(channel Channel, id ora.ScopeID) *Scope {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(id) < 32 {
		id = ora.NewScopeID()
	}

	scope := a.scopes[id]
	if scope == nil {
		scope = NewScope(id, time.Minute, a.factories)
	}

	scope.Connect(channel)
	a.scopes[id] = scope

	return scope
}
