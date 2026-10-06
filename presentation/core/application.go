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
	"go.wdy.de/nago/logging"
	"io"
	"log/slog"
	"maps"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob/crypto"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/presentation/proto"
)

var appIdRegex = regexp.MustCompile(`^[a-z]\w*(\.[a-z]\w*)+$`)

type ApplicationID string

func (a ApplicationID) Valid() bool {
	return appIdRegex.FindString(string(a)) == string(a)
}

func (a ApplicationID) Last() string {
	tokens := strings.Split(string(a), ".")
	return tokens[len(tokens)-1]
}

type OnWindowCreatedObserver func(wnd Window)

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
	colorsMutex              sync.RWMutex // protects colorSets, which windows read while the theme may change
	colorSets                map[ColorScheme]map[NamespaceName]ColorSet

	findVirtualSession session.FindUserSessionByID

	masterKey     crypto.EncryptionKey
	bus           events.Bus
	getAnonUser   user.GetAnonUser
	logoutSession session.Logout
	fonts         atomic.Pointer[Fonts]
	debug         bool
	instance      string
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
	getAnonUser user.GetAnonUser,
	logoutSession session.Logout,
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
		bus:           bus,
		getAnonUser:   getAnonUser,
		logoutSession: logoutSession,
		instance:      data.RandIdent[string](),
	}

	return a
}

// Instance returns a random identifier for the application instance which is created on application startup
// and is different for each created instance. It can be used by any client to detect if the service has been
// restarted, which may also indicate that the protocol may have been changed and any frontend
// JavaScript may need a reload.
func (a *Application) Instance() string {
	return a.instance
}

// Context of the application.
func (a *Application) Context() context.Context {
	a.mutex.Lock()
	defer a.mutex.Unlock()

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

func (a *Application) EventBus() events.Bus {
	return a.bus
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

func (a *Application) IsDebug() bool {
	return a.debug
}

func (a *Application) SetDebug(debug bool) {
	a.debug = debug
}

func (a *Application) SetVersion(version string) {
	a.version = version
}

func (a *Application) SetAppIcon(appIcon URI) {
	a.appIcon = appIcon
}

// UpdateColorSet replaces the color set of its namespace for the given scheme. It may be called from any
// goroutine, also while windows render.
func (a *Application) UpdateColorSet(scheme ColorScheme, set ColorSet) {
	a.colorsMutex.Lock()
	defer a.colorsMutex.Unlock()

	sets, ok := a.colorSets[scheme]
	if !ok {
		sets = map[NamespaceName]ColorSet{}
		a.colorSets[scheme] = sets
	}

	sets[set.Namespace()] = set
}

// colorSetsSnapshot returns a copy of all color sets.
func (a *Application) colorSetsSnapshot() map[ColorScheme]map[NamespaceName]ColorSet {
	a.colorsMutex.RLock()
	defer a.colorsMutex.RUnlock()

	res := make(map[ColorScheme]map[NamespaceName]ColorSet, len(a.colorSets))
	for scheme, sets := range a.colorSets {
		res[scheme] = maps.Clone(sets)
	}

	return res
}

// colorSet returns the color set of the namespace for the given scheme. knownScheme is false, if there is no
// color set at all for the scheme.
func (a *Application) colorSet(scheme ColorScheme, ns NamespaceName) (set ColorSet, knownScheme bool, found bool) {
	a.colorsMutex.RLock()
	defer a.colorsMutex.RUnlock()

	sets, knownScheme := a.colorSets[scheme]
	if !knownScheme {
		return nil, false, false
	}

	set, found = sets[ns]
	return set, true, found
}

func (a *Application) UpdateFonts(fonts Fonts) {
	a.fonts.Store(&fonts)
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

// AddOnWindowCreatedObserver appends the observer, which is called after the already configured observers
// whenever a window is created. It must be called before any scope has been connected.
func (a *Application) AddOnWindowCreatedObserver(observer OnWindowCreatedObserver) {
	a.onWindowCreatedObservers = append(a.onWindowCreatedObservers, observer)
}

func (a *Application) Scope(id proto.ScopeID) (*Scope, bool) {
	return a.scopes.Get(id)
}

// DestroyScope removes the scope of the given id and destroys it. It returns false, if no such scope exists.
func (a *Application) DestroyScope(id proto.ScopeID) bool {
	return a.scopes.Remove(id)
}

// Connect either connects an existing scope with the channel or creates a new scope with the given id.
func (a *Application) Connect(channel Channel, id proto.ScopeID) *Scope {
	a.mutex.Lock()
	// protect only the moment of connecting against races. perhaps it may be even ok, to remove the entire lock
	// add say that concurrent calls to the same scope id is invalid (and normally cannot happen)

	if len(id) < 32 {
		id = proto.NewScopeID()
	}

	scope := a.scopes.getOrCreate(id, func() *Scope {
		return NewScope(a.ctx, a, filepath.Join(a.tmpDir, string(id)), id, time.Minute, a.factories, a.findVirtualSession)
	})

	a.mutex.Unlock()

	scope.Connect(channel)
	return scope
}

func (a *Application) ImportFilesOptions(scopeId proto.ScopeID, uploadId string) (ImportFilesOptions, bool) {
	scope, ok := a.scopes.Get(scopeId)
	if !ok {
		slog.Error("no such scope to import files", "scope", logging.Secret(string(scopeId)))
		return ImportFilesOptions{}, false
	}

	return scope.ImportFilesOptions(uploadId)
}

func (a *Application) ExportFilesOptions(scopeId proto.ScopeID, downloadId string) (ExportFilesOptions, bool) {
	scope, ok := a.scopes.Get(scopeId)
	if !ok {
		slog.Error("no such scope to export files", "scope", logging.Secret(string(scopeId)))
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
