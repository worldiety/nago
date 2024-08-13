package core

import (
	"context"
	"fmt"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/presentation/ora"
	"io"
	"log/slog"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

var appIdRegex = regexp.MustCompile(`^[a-z]\w*(\.[a-z]\w*)+$`)

type ApplicationID string

func (a ApplicationID) Valid() bool {
	return appIdRegex.FindString(string(a)) == string(a)
}

type OnWindowCreatedObserver func(wnd Window)

type Application struct {
	id                       ApplicationID
	name                     string
	version                  string
	appIcon                  ora.URI
	mutex                    sync.Mutex
	scopes                   *Scopes
	factories                map[ora.ComponentFactoryId]ComponentFactory
	scopeLifetime            time.Duration
	ctx                      context.Context
	cancelCtx                func()
	tmpDir                   string
	onSendFiles              func(scope *Scope, options ExportFilesOptions) error
	onShareStream            func(*Scope, func() (io.Reader, error)) (URI, error)
	onWindowCreatedObservers []OnWindowCreatedObserver
	destructors              *concurrent.LinkedList[func()]
	colorSets                map[ColorScheme]map[NamespaceName]ColorSet
}

func NewApplication(ctx context.Context, tmpDir string, factories map[ora.ComponentFactoryId]ComponentFactory, onWindowCreatedObservers []OnWindowCreatedObserver, fps int) *Application {
	cancelCtx, cancel := context.WithCancel(ctx)

	return &Application{
		destructors:              concurrent.NewLinkedList[func()](),
		scopeLifetime:            time.Minute,
		factories:                factories,
		scopes:                   NewScopes(fps),
		ctx:                      cancelCtx,
		cancelCtx:                cancel,
		tmpDir:                   tmpDir,
		onWindowCreatedObservers: onWindowCreatedObservers,
		colorSets: map[ColorScheme]map[NamespaceName]ColorSet{
			Light: {},
			Dark:  {},
		},
	}
}

func (a *Application) SetID(id ApplicationID) {
	if !id.Valid() {
		panic(fmt.Errorf("invalid application id"))
	}
	a.id = id
}

func (a *Application) SetName(name string) {
	a.name = name
}

func (a *Application) Version() string {
	return a.version
}

func (a *Application) Name() string {
	return a.name
}

func (a *Application) ID() ApplicationID {
	return a.id
}

func (a *Application) SetVersion(version string) {
	a.version = version
}

func (a *Application) SetAppIcon(appIcon ora.URI) {
	a.appIcon = appIcon
}

func (a *Application) AddColorSet(scheme ColorScheme, set ColorSet) {
	a.colorSets[scheme][set.Namespace()] = set
}

// SetOnSendFiles sets the callback which is called by the window or application to trigger the platform specific
// "send files" behavior. On webbrowser the according download events may be issued and on other platforms
// like Android a custom content provider may be created which exposes these blobs as URIs.
func (a *Application) SetOnSendFiles(onSendFiles func(scope *Scope, options ExportFilesOptions) error) {
	a.onSendFiles = onSendFiles
}

// SetOnShareStream set the callback which is called the by the window to convert any dynamic stream into a fixed
// URI. A webbrowser will get an url resource, which must not be cached. Android needs a custom content provider.
func (a *Application) SetOnShareStream(onShareStream func(*Scope, func() (io.Reader, error)) (URI, error)) {
	a.onShareStream = onShareStream
}

func (a *Application) Scope(id ora.ScopeID) (*Scope, bool) {
	return a.scopes.Get(id)
}

// Connect either connects an existing scope with the channel or creates a new scope with the given id.
func (a *Application) Connect(channel Channel, id ora.ScopeID) *Scope {
	a.mutex.Lock()
	// protect only the moment of connecting against races. perhaps it may be even ok, to remove the entire lock
	// add say that concurrent calls to the same scope id is invalid (and normally cannot happen)

	if len(id) < 32 {
		id = ora.NewScopeID()
	}

	scope, _ := a.scopes.Get(id)
	if scope == nil {
		scope = NewScope(a.ctx, a, filepath.Join(a.tmpDir, string(id)), id, time.Minute, a.factories)
	}
	a.scopes.Put(scope)

	a.mutex.Unlock()

	scope.Connect(channel)
	return scope
}

func (a *Application) ImportFilesOptions(scopeId ora.ScopeID, uploadId string) (ImportFilesOptions, bool) {
	scope, ok := a.scopes.Get(scopeId)
	if !ok {
		slog.Error("no such scope to import files", "scope", scope.id)
		return ImportFilesOptions{}, false
	}

	return scope.ImportFilesOptions(uploadId)
}

func (a *Application) ExportFilesOptions(scopeId ora.ScopeID, downloadId string) (ExportFilesOptions, bool) {
	scope, ok := a.scopes.Get(scopeId)
	if !ok {
		slog.Error("no such scope to export files", "scope", scope.id)
		return ExportFilesOptions{}, false
	}

	return scope.ExportFilesOptions(downloadId)
}

func (a *Application) AddDestructor(f func()) {
	a.destructors.PushBack(f)
}

func (a *Application) Destroy() {
	//a.mutex.Lock() probably unneeded locks
	//defer a.mutex.Unlock()

	for _, destructor := range a.destructors.Values() {
		destructor()
	}
	a.destructors.Clear()

	a.scopes.Destroy()
	a.cancelCtx()
}
