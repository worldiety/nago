// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// SettleTimeout is the maximum duration [Window.Settle] waits for the window to become idle, before the
// test fails.
var SettleTimeout = 10 * time.Second

// Route is a navigation target of a window.
type Route struct {
	Path   core.NavigationPath
	Values core.Values
}

// OpenOption configures a [Window].
type OpenOption func(*openOptions)

type openOptions struct {
	values     core.Values
	locale     string
	windowInfo proto.WindowInfo
}

// Values sets the query values of the initial route.
func Values(values core.Values) OpenOption {
	return func(o *openOptions) {
		o.values = values
	}
}

// Locale sets the accepted language of the window, e.g. "de" or "en". The default is "de".
func Locale(locale string) OpenOption {
	return func(o *openOptions) {
		o.locale = locale
	}
}

// Size sets the window size in dp. The size class is derived from the width. The default is 1280x800.
func Size(width, height int) OpenOption {
	return func(o *openOptions) {
		o.windowInfo.Width = proto.DP(width)
		o.windowInfo.Height = proto.DP(height)
		o.windowInfo.SizeClass = sizeClassOf(width)
	}
}

func sizeClassOf(width int) proto.WindowSizeClass {
	// see core.WindowSizeClass breakpoints
	switch {
	case width < 640:
		return proto.SizeClassSmall
	case width < 768:
		return proto.SizeClassMedium
	case width < 1024:
		return proto.SizeClassLarge
	case width < 1280:
		return proto.SizeClassXL
	default:
		return proto.SizeClass2XL
	}
}

// A Window is a browser window, which is connected to the [App] within the same process. It emulates the
// frontend: it keeps the latest rendered tree, follows navigation requests and records all other side
// effects. All methods must be called from the test goroutine.
type Window struct {
	t      testing.TB
	app    *App
	scope  *core.Scope
	opts   openOptions
	rid    proto.RID
	closed bool

	mutex      sync.Mutex
	tree       proto.Component
	history    []Route
	events     []proto.NagoEvent
	calls      []*proto.CallRequested
	downloads  []core.ExportFilesOptions
	imports    []*proto.FileImportRequested
	opened     []Route
	links      []string
	focus      string
	clipboard  string
	errs       []string
	listeners  map[proto.Uint]listener
	canvasSeen map[string]int
	navigation []Route // pending navigations requested by the application
}

// Open connects a new window as the given subject and allocates the root view of the given path, just like a
// browser which opens the according URL. If subject is nil, the window is anonymous.
// The returned window has already settled and is closed automatically at the end of the test.
func (a *App) Open(t testing.TB, subject auth.Subject, path core.NavigationPath, opts ...OpenOption) *Window {
	t.Helper()

	w := &Window{t: t, app: a}
	w.opts.locale = "de"
	w.opts.windowInfo = proto.WindowInfo{Density: 2, ColorScheme: proto.Light}
	Size(1280, 800)(&w.opts)
	for _, opt := range opts {
		opt(&w.opts)
	}

	sessionID := data.RandIdent[string]()
	if subject != nil {
		a.subjects.Put(session.ID(sessionID), subject)
	}

	w.scope = a.core.Connect(&channel{w: w}, proto.NewScopeID())
	a.windows.Put(w.scope.ID(), w)
	t.Cleanup(w.Close)

	w.dispatch(&proto.ScopeConfigurationChangeRequested{
		AcceptLanguage: proto.Locale(w.opts.locale),
		WindowInfo:     w.opts.windowInfo,
		RID:            w.nextRID(),
	})
	w.dispatch(&proto.SessionAssigned{SessionID: proto.Str(sessionID)})

	w.history = []Route{{Path: path, Values: w.opts.values}}
	w.allocate(w.history[0])
	w.Settle()

	return w
}

func (w *Window) nextRID() proto.RID {
	w.rid++
	return w.rid
}

func (w *Window) dispatch(evt proto.NagoEvent) {
	w.t.Helper()

	if w.app.opts.roundTrip {
		evt = roundTrip(w.t, evt)
	}

	if err := w.scope.Dispatch(evt); err != nil {
		w.t.Fatalf("nagotest: cannot dispatch %T: %v", evt, err)
	}
}

func (w *Window) allocate(route Route) {
	w.t.Helper()

	values := proto.RootViewParameters{}
	for k, v := range route.Values {
		values[proto.Str(k)] = proto.Str(v)
	}

	w.dispatch(&proto.RootViewAllocationRequested{
		Locale:  proto.Locale(w.opts.locale),
		Factory: proto.RootViewID(route.Path),
		RID:     w.nextRID(),
		Values:  values,
	})
}

// Settle waits until the window is idle: all events have been processed, pending state changes have been
// rendered, background tasks like [core.OnAppear] have completed and requested navigations have been
// followed. Delayed functions (see [core.Window.PostDelayed]) are not awaited. All actions of a Window
// settle automatically, thus you only need to call this after changing application state from outside.
func (w *Window) Settle() {
	w.t.Helper()

	deadline := time.Now().Add(SettleTimeout)
	for {
		if w.scope.Flush() && !w.followNavigation() {
			break
		}

		if time.Now().After(deadline) {
			w.t.Fatalf("nagotest: window did not settle within %v", SettleTimeout)
		}

		// only background goroutines can keep us busy without posting to the event loop, thus don't spin
		time.Sleep(time.Millisecond)
	}

	w.mutex.Lock()
	errs := w.errs
	w.errs = nil
	w.mutex.Unlock()

	for _, err := range errs {
		w.t.Errorf("nagotest: %s", err)
	}
}

// followNavigation emulates the browser history and allocates the root view of the navigation requested last.
// It returns true, if a navigation has been followed.
func (w *Window) followNavigation() bool {
	w.mutex.Lock()
	nav := w.navigation
	w.navigation = nil
	w.mutex.Unlock()

	if len(nav) == 0 {
		return false
	}

	for _, route := range nav {
		w.allocate(route)
	}

	return true
}

// Close destroys the window immediately, like closing the browser tab. Closing twice has no effect.
func (w *Window) Close() {
	if w.closed {
		return
	}

	w.closed = true
	w.scope.Destroy()
	w.scope.Flush()
	w.app.windows.Delete(w.scope.ID())
}

// Reload emulates a browser reload: the root view is destroyed and allocated again, thus all window
// states are lost. The scope and the session are kept.
func (w *Window) Reload() {
	w.t.Helper()

	w.dispatch(&proto.RootViewDestructionRequested{RID: w.nextRID()})
	w.allocate(w.Route())
	w.Settle()
}

// Reconnect emulates a lost connection: a new channel is connected to the existing scope, which is configured
// and requests the current root view again, as the frontend does. The window states are kept.
func (w *Window) Reconnect() {
	w.t.Helper()

	w.scope = w.app.core.Connect(&channel{w: w}, w.scope.ID())
	w.dispatch(&proto.ScopeConfigurationChangeRequested{
		AcceptLanguage: proto.Locale(w.opts.locale),
		WindowInfo:     w.opts.windowInfo,
		RID:            w.nextRID(),
	})
	w.allocate(w.Route())
	w.Settle()
}

// Scope returns the underlying scope.
func (w *Window) Scope() *core.Scope {
	return w.scope
}

// Tree returns the root of the most recently rendered tree.
func (w *Window) Tree() proto.Component {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.tree
}

// Route returns the current route of the window.
func (w *Window) Route() Route {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.history[len(w.history)-1]
}

// History returns the navigation history of the window, the current route is last.
func (w *Window) History() []Route {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return append([]Route(nil), w.history...)
}

// Opened returns the routes which the application requested to open in another window or tab
// (see [core.Navigation.ForwardToTarget]).
func (w *Window) Opened() []Route {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return append([]Route(nil), w.opened...)
}

// Links returns the external links or flows which the application requested to open
// (see [core.Navigation.Open]).
func (w *Window) Links() []string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return append([]string(nil), w.links...)
}

// Downloads returns all files which have been sent using [core.Window.ExportFiles].
func (w *Window) Downloads() []core.ExportFilesOptions {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return append([]core.ExportFilesOptions(nil), w.downloads...)
}

// Imports returns all file import requests issued using [core.Window.ImportFiles].
func (w *Window) Imports() []*proto.FileImportRequested {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return append([]*proto.FileImportRequested(nil), w.imports...)
}

// AsyncCalls returns all frontend calls requested using [core.AsyncCall], e.g. canvas commands or
// input listener registrations.
func (w *Window) AsyncCalls() []*proto.CallRequested {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return append([]*proto.CallRequested(nil), w.calls...)
}

// Focused returns the component ID which requested the focus last, see [core.Window.RequestFocus].
func (w *Window) Focused() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.focus
}

// Clipboard returns the text written last into the clipboard.
func (w *Window) Clipboard() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.clipboard
}

// Events returns all events published by the application to this window.
func (w *Window) Events() []proto.NagoEvent {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return append([]proto.NagoEvent(nil), w.events...)
}

// receive is called from the event loop of the scope.
func (w *Window) receive(evt proto.NagoEvent) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.events = append(w.events, evt)

	switch evt := evt.(type) {
	case *proto.RootViewInvalidated:
		w.tree = evt.Root
	case *proto.NavigationForwardToRequested:
		route := routeOf(evt.RootView, evt.Values)
		if evt.Target != "" && evt.Target != "_self" {
			w.opened = append(w.opened, route)
			return
		}

		w.history = append(w.history, route)
		w.navigation = append(w.navigation, route)
	case *proto.NavigationBackRequested:
		if len(w.history) > 1 {
			w.history = w.history[:len(w.history)-1]
		}
		w.navigation = append(w.navigation, w.history[len(w.history)-1])
	case *proto.NavigationReplaceRequested:
		route := routeOf(evt.RootView, evt.Values)
		w.history[len(w.history)-1] = route
		w.navigation = append(w.navigation, route)
	case *proto.NavigationResetRequested:
		route := routeOf(evt.RootView, evt.Values)
		w.history = []Route{route}
		w.navigation = append(w.navigation, route)
	case *proto.NavigationReloadRequested:
		w.navigation = append(w.navigation, w.history[len(w.history)-1])
	case *proto.OpenHttpLink:
		w.links = append(w.links, string(evt.Url))
	case *proto.OpenHttpFlow:
		w.links = append(w.links, string(evt.Url))
	case *proto.FileImportRequested:
		w.imports = append(w.imports, evt)
	case *proto.CallRequested:
		w.calls = append(w.calls, evt)
		switch call := evt.Call.(type) {
		case *proto.CallRequestFocus:
			w.focus = string(call.ID)
		case *proto.RegisterInputEventListener:
			if w.listeners == nil {
				w.listeners = map[proto.Uint]listener{}
			}
			w.listeners[call.Handle] = listener{ptr: evt.CallPtr, call: call}
		case *proto.UnregisterInputEventListener:
			delete(w.listeners, call.Handle)
		}
	case *proto.ClipboardWriteTextRequested:
		w.clipboard = string(evt.Text)
	case *proto.ErrorOccurred:
		w.errs = append(w.errs, string(evt.Message))
	case *proto.ErrorRootViewAllocationRequired:
		w.errs = append(w.errs, "root view allocation required")
	}
}

func routeOf(id proto.RootViewID, params proto.RootViewParameters) Route {
	values := core.Values{}
	for k, v := range params {
		values[string(k)] = string(v)
	}

	return Route{Path: core.NavigationPath(id), Values: values}
}

// channel connects a scope with a window without encoding.
type channel struct {
	w *Window
}

func (c *channel) Subscribe(f func(msg []byte) error) (destroy func()) {
	return func() {}
}

func (c *channel) Publish(msg []byte) error {
	obj, err := proto.Unmarshal(proto.NewBinaryReader(bytes.NewBuffer(msg)))
	if err != nil {
		return err
	}

	evt, ok := obj.(proto.NagoEvent)
	if !ok {
		return fmt.Errorf("nagotest: %T is not a NagoEvent", obj)
	}

	return c.PublishEvent(evt)
}

func (c *channel) PublishEvent(evt proto.NagoEvent) error {
	if c.w.app.opts.roundTrip {
		evt = roundTrip(c.w.t, evt)
	}

	c.w.receive(evt)
	return nil
}

func roundTrip(t testing.TB, evt proto.NagoEvent) proto.NagoEvent {
	var buf bytes.Buffer
	if err := proto.Marshal(proto.NewBinaryWriter(&buf), evt); err != nil {
		t.Errorf("nagotest: cannot encode %T: %v", evt, err)
		return evt
	}

	obj, err := proto.Unmarshal(proto.NewBinaryReader(&buf))
	if err != nil {
		t.Errorf("nagotest: cannot decode %T: %v", evt, err)
		return evt
	}

	return obj.(proto.NagoEvent)
}
