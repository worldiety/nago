package core

import (
	"context"
	"fmt"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/ora"
	"io"
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
	onSendFiles   func(*Scope, iter.Seq2[File, error]) error
	onShareStream func(*Scope, func() (io.Reader, error)) (ora.URI, error)
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

// SetOnSendFiles sets the callback which is called by the window or application to trigger the platform specific
// "send files" behavior. On webbrowser the according download events may be issued and on other platforms
// like Android a custom content provider may be created which exposes these blobs as URIs.
func (a *Application) SetOnSendFiles(onSendFiles func(*Scope, iter.Seq2[File, error]) error) {
	a.onSendFiles = onSendFiles
}

// SetOnShareStream set the callback which is called the by the window to convert any dynamic stream into a fixed
// URI. A webbrowser will get an url resource, which must not be cached. Android needs a custom content provider.
func (a *Application) SetOnShareStream(onShareStream func(*Scope, func() (io.Reader, error)) (ora.URI, error)) {
	a.onShareStream = onShareStream
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
		scope = NewScope(a.ctx, a, filepath.Join(a.tmpDir, string(id)), id, time.Minute, a.factories)
	}

	scope.Connect(channel)
	a.scopes.Put(scope)

	return scope
}

// OnFilesReceived delegates the received fs into according scope.
func (a *Application) OnFilesReceived(scopeId ora.ScopeID, receiver ora.Ptr, it iter.Seq2[File, error]) error {
	scope, ok := a.scopes.Get(scopeId)
	if !ok {
		return fmt.Errorf("no such scope to receive stream: %s", scope.id)
	}

	return scope.OnFilesReceived(receiver, it)
}

func (a *Application) Destroy() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.scopes.Destroy()
	a.cancelCtx()
}
