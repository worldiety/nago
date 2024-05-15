package core

import (
	"context"
	"fmt"
	"go.wdy.de/nago/presentation/ora"
	"io/fs"
	"path/filepath"
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
	tmpDir        string
}

func NewApplication(ctx context.Context, tmpDir string, factories map[ora.ComponentFactoryId]ComponentFactory) *Application {
	cancelCtx, cancel := context.WithCancel(ctx)

	return &Application{
		scopeLifetime: time.Minute,
		factories:     factories,
		scopes:        NewScopes(),
		ctx:           cancelCtx,
		cancelCtx:     cancel,
		tmpDir:        tmpDir,
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
		scope = NewScope(a.ctx, filepath.Join(a.tmpDir, string(id)), id, time.Minute, a.factories)
	}

	scope.Connect(channel)
	a.scopes.Put(scope)

	return scope
}

// OnFilesReceived delegates the received fs into according scope.
func (a *Application) OnFilesReceived(scopeId ora.ScopeID, receiver ora.Ptr, fsys fs.FS) error {
	scope, ok := a.scopes.Get(scopeId)
	if !ok {
		return fmt.Errorf("no such scope to receive stream: %s", scope.id)
	}

	return scope.OnFilesReceived(receiver, fsys)
}

func (a *Application) Destroy() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.scopes.Destroy()
	a.cancelCtx()
}
