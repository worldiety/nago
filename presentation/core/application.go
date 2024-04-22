package core

import (
	"context"
	"go.wdy.de/nago/presentation/ora"
	"sync"
	"time"
)

type Application struct {
	mutex         sync.Mutex
	scopes        *Scopes
	factories     map[ora.ComponentFactoryId]ComponentFactory
	scopeLifetime time.Duration
	ctx           context.Context
	cancelCtx     func()
}

func NewApplication(ctx context.Context, factories map[ora.ComponentFactoryId]ComponentFactory) *Application {
	cancelCtx, cancel := context.WithCancel(ctx)

	return &Application{
		scopeLifetime: time.Minute,
		factories:     factories,
		scopes:        NewScopes(),
		ctx:           cancelCtx,
		cancelCtx:     cancel,
	}
}

// Connect either connects an existing scope with the channel or creates a new scope with the given id.
func (a *Application) Connect(channel Channel, id ora.ScopeID) *Scope {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(id) < 32 {
		id = ora.NewScopeID()
	}

	scope, _ := a.scopes.Get(id)
	if scope == nil {
		scope = NewScope(a.ctx, id, time.Minute, a.factories)
	}

	scope.Connect(channel)
	a.scopes.Put(scope)

	return scope
}

func (a *Application) Destroy() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.scopes.Destroy()
	a.cancelCtx()
}
