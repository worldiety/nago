package core

import (
	"go.wdy.de/nago/presentation/protocol"
	"sync"
	"time"
)

type Application struct {
	mutex         sync.Mutex
	scopes        map[protocol.ScopeID]*Scope
	factories     map[protocol.ComponentFactoryId]ComponentFactory
	scopeLifetime time.Duration
}

func NewApplication(factories map[protocol.ComponentFactoryId]ComponentFactory) *Application {
	return &Application{
		scopeLifetime: time.Minute,
		factories:     factories,
		scopes:        map[protocol.ScopeID]*Scope{},
	}
}

// Connect either connects an existing scope with the channel or creates a new scope with the given id.
func (a *Application) Connect(channel Channel, id protocol.ScopeID) *Scope {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(id) < 32 {
		id = protocol.NewScopeID()
	}

	scope := a.scopes[id]
	if scope == nil {
		scope = NewScope(id, time.Minute, a.factories)
	}

	scope.Connect(channel)
	a.scopes[id] = scope

	return scope
}
