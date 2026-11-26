// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"bytes"
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
	onSendFiles              func(scope *Scope, token string, options ExportFilesOptions) error
	onShareStream            func(*Scope, func() (io.Reader, error)) (URI, error)
	downloads                downloads
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

	if bus != nil {
		a.AddDestructor(events.SubscribeFor[session.LoggedOut](bus, a.sessionLoggedOut))
	}

	return a
}

// sessionLoggedOut shows the open windows of the session as logged out, e.g. after a logout in another tab or a
// failed single sign-on refresh in the background. Otherwise, a window keeps its subject until its next navigation.
func (a *Application) sessionLoggedOut(evt session.LoggedOut) {
	a.scopes.scopes.Each(func(_ proto.ScopeID, scope *Scope) bool {
		scope.sessionLoggedOut(evt.Session)
		return true
	})
}

// PublishConfiguration sends the current themes and fonts to all open windows, which apply them at once. Call it
// after changing them with [Application.UpdateColorSet] or [Application.UpdateFonts], otherwise a window shows them
// only after a reload.
func (a *Application) PublishConfiguration() {
	a.scopes.scopes.Each(func(_ proto.ScopeID, scope *Scope) bool {
		scope.eventLoop.Post(func() {
			scope.publishConfiguration(0)
		})
		return true
	})
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
// like Android a custom content provider may be created which exposes these blobs as URIs. The token names the
// download for [Application.Download], which delivers it only to the owner of the window, see [LinkOwner].
func (a *Application) SetOnSendFiles(onSendFiles func(scope *Scope, token string, options ExportFilesOptions) error) {
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

// Connect either connects an existing scope with the channel or creates a new scope with the given id. It is for a
// transport without a session of its own, e.g. within the same process, whose client assigns its session by a
// message; a transport, which authenticates the session, e.g. by a cookie, uses [Application.ConnectSession].
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

// ConnectSession connects the channel of a transport, which authenticated the session, e.g. by the http-only
// cookie of the wire, to the scope of the given id and assigns the session. A new id creates a scope, which the
// session owns from then on. The scope of an id, which another session owns, is refused: it reports false and
// leaves the scope untouched, because the id of a scope is part of the address of the wire and thus no secret
// enough to grant its window, which shows its subject and state. A transport tells the client of a refused
// channel to load the page again, see [RequestReload], which gives it a scope of its own.
func (a *Application) ConnectSession(channel Channel, id proto.ScopeID, sid session.ID) (*Scope, bool) {
	a.mutex.Lock()
	if len(id) < 32 {
		id = proto.NewScopeID()
	}

	scope, ok := a.scopes.connectOwned(id, sid, func() *Scope {
		return NewScope(a.ctx, a, filepath.Join(a.tmpDir, string(id)), id, time.Minute, a.factories, a.findVirtualSession)
	})
	a.mutex.Unlock()

	if !ok {
		slog.Warn("refused a connection of another session to a scope", "scope", logging.Secret(string(id)))
		return nil, false
	}

	scope.Connect(channel)
	scope.AssignSession(sid)
	return scope, true
}

// RequestReload asks the frontend of the channel to load its page again, e.g. after [Application.ConnectSession]
// refused the channel. A browser then starts a new window with a new scope id, instead of retrying the refused one.
func RequestReload(c Channel) error {
	buf := bytes.NewBuffer(make([]byte, 0, 16))
	if err := proto.Marshal(proto.NewBinaryWriter(buf), &proto.NavigationReloadRequested{}); err != nil {
		return err
	}

	return c.Publish(buf.Bytes())
}

// ImportFilesOptions returns the pending upload of the scope with the given id, if the requesting session owns
// the scope, see [Application.ConnectSession]. Otherwise, it reports false like for an unknown upload, without
// keeping the scope alive. A scope connected without a session accepts no upload of a request.
func (a *Application) ImportFilesOptions(scopeId proto.ScopeID, uploadId string, requester session.ID) (ImportFilesOptions, bool) {
	scope, ok := a.scopes.Get(scopeId)
	if !ok {
		slog.Error("no such scope to import files", "scope", logging.Secret(string(scopeId)))
		return ImportFilesOptions{}, false
	}

	if !scope.ownedBy(requester) {
		slog.Warn("refused an upload of another session", "scope", logging.Secret(string(scopeId)))
		return ImportFilesOptions{}, false
	}

	return scope.ImportFilesOptions(uploadId)
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
