// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"context"
	"fmt"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/pkg/blob/crypto"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/presentation/proto"
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

type dependency struct {
	name  string
	value any
}

type Application struct {
	id                       ApplicationID
	name                     string
	version                  string
	appIcon                  URI
	mutex                    sync.Mutex
	scopes                   *Scopes
	factories                map[proto.RootViewID]ComponentFactory
	scopeLifetime            time.Duration
	ctx                      context.Context
	cancelCtx                func()
	tmpDir                   string
	onSendFiles              func(scope *Scope, options ExportFilesOptions) error
	onShareStream            func(*Scope, func() (io.Reader, error)) (URI, error)
	onWindowCreatedObservers []OnWindowCreatedObserver
	destructors              *concurrent.LinkedList[func()]
	colorSets                map[ColorScheme]map[NamespaceName]ColorSet
	// this fills a dependency hole just the same way as the android system service call.
	// sometimes we have just a deeply nested thing and we just want eventually access to
	// a global instance of something. I actually never wanted this kind of API
	// but any other API does not make sense for things like crud.Auto where a lot of stuff
	// is only eventually relevant and available automatically anyway.
	systemServices []dependency

	findVirtualSession session.FindUserSessionByID

	masterKey crypto.EncryptionKey
	bus       events.Bus
}

func NewApplication(
	ctx context.Context,
	tmpDir string,
	factories map[proto.RootViewID]ComponentFactory,
	onWindowCreatedObservers []OnWindowCreatedObserver,
	fps int,
	findVirtualSession session.FindUserSessionByID,
	masterKey crypto.EncryptionKey,
	bus events.Bus,
) *Application {
	cancelCtx, cancel := context.WithCancel(ctx)

	a := &Application{
		masterKey:                masterKey,
		findVirtualSession:       findVirtualSession,
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
		bus: bus,
	}

	return a
}

// Context of the application.
func (a *Application) Context() context.Context {
	return a.ctx
}

func (a *Application) MasterKey() crypto.EncryptionKey {
	return a.masterKey
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

func (a *Application) SetAppIcon(appIcon URI) {
	a.appIcon = appIcon
}

func (a *Application) UpdateColorSet(scheme ColorScheme, set ColorSet) {
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

func (a *Application) Scope(id proto.ScopeID) (*Scope, bool) {
	return a.scopes.Get(id)
}

// Connect either connects an existing scope with the channel or creates a new scope with the given id.
func (a *Application) Connect(channel Channel, id proto.ScopeID) *Scope {
	a.mutex.Lock()
	// protect only the moment of connecting against races. perhaps it may be even ok, to remove the entire lock
	// add say that concurrent calls to the same scope id is invalid (and normally cannot happen)

	if len(id) < 32 {
		id = proto.NewScopeID()
	}

	scope, _ := a.scopes.Get(id)
	if scope == nil {
		scope = NewScope(a.ctx, a, filepath.Join(a.tmpDir, string(id)), id, time.Minute, a.factories, a.findVirtualSession)
	}
	a.scopes.Put(scope)

	a.mutex.Unlock()

	scope.Connect(channel)
	return scope
}

func (a *Application) ImportFilesOptions(scopeId proto.ScopeID, uploadId string) (ImportFilesOptions, bool) {
	scope, ok := a.scopes.Get(scopeId)
	if !ok {
		slog.Error("no such scope to import files", "scope", scope.id)
		return ImportFilesOptions{}, false
	}

	return scope.ImportFilesOptions(uploadId)
}

func (a *Application) ExportFilesOptions(scopeId proto.ScopeID, downloadId string) (ExportFilesOptions, bool) {
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

// AddSystemService appends the given service to the internal list. It is kept alive and
// accessible through [SystemService]. If another service of the same type has already been added,
// the newly added service has a higher precedence but the older service is not removed.
func (a *Application) AddSystemService(service any) {
	a.AddSystemServiceWithName("", service)
}

// AddSystemServiceWithName is like [AddSystemService], but adds a name as an additional criteria.
func (a *Application) AddSystemServiceWithName(name string, service any) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.systemServices = append(a.systemServices, dependency{
		name:  name,
		value: service,
	})
}

// SystemServiceByName returns last added service identified by the given name.
func (a *Application) SystemServiceByName(name string) (any, bool) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for i := len(a.systemServices) - 1; i >= 0; i-- {
		service := a.systemServices[i]
		if service.name == name {
			return service.value, true
		}
	}
	return nil, false
}

// SystemService returns the T which is assignable to any of the registered Scheduler via
// [Application.AddSystemService] or returns false. The matching is applied reverse, so that the last added type
// is matched first. The name is ignored. See also [SystemServiceWithName].
func SystemService[T any](app *Application) (T, bool) {
	return SystemServiceWithName[T](app, "")
}

// SystemServiceWithName is like [SystemService] but filters by name. The last added named and matching service
// has the highest precedence.
func SystemServiceWithName[T any](app *Application, name string) (T, bool) {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	for i := len(app.systemServices) - 1; i >= 0; i-- {
		service := app.systemServices[i]
		if name != "" && service.name != name {
			continue
		}

		if t, ok := service.value.(T); ok {
			return t, ok
		}
	}

	var zero T

	return zero, false
}
