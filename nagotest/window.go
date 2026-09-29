// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"fmt"
	"sync"
	"time"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// SettleTimeout is the maximum duration [Window.Settle] waits for the window to become idle, before the
// test fails.
var SettleTimeout = 10 * time.Second

// TB is the subset of [testing.TB] used by a [Window]. Besides tests, it is implemented by the virtual users
// of the load package. Fatalf must not return, e.g. by calling [runtime.Goexit].
type TB interface {
	Helper()
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Logf(format string, args ...any)
	Cleanup(func())
}

// Route is a navigation target of a window.
type Route struct {
	Path   core.NavigationPath
	Values core.Values
}

// An Action describes a completed user interaction of a [Window], including the time until the window
// settled. See [Observe].
type Action struct {
	// Kind is one of open, click, type, enter, input, upload, reload, reconnect or wait.
	Kind string
	// Target describes the path or the selection, e.g. Text("Speichern").
	Target   string
	Duration time.Duration
}

// OpenOption configures a [Window].
type OpenOption func(*openOptions)

type openOptions struct {
	values     core.Values
	locale     string
	windowInfo proto.WindowInfo
	sessionID  string
	observer   func(Action)
	lean       bool
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

// Session sets the session ID, which must have at least 32 characters. Windows with the same session share
// the login state, like tabs of the same browser. By default, each window has its own session.
func Session(id string) OpenOption {
	return func(o *openOptions) {
		o.sessionID = id
	}
}

// Observe calls fn after each completed action, e.g. to measure latencies.
func Observe(fn func(Action)) OpenOption {
	return func(o *openOptions) {
		o.observer = fn
	}
}

// Lean keeps only the current state of the window instead of the entire history of received events and
// asynchronous calls. [Window.Events], [Window.AsyncCalls] and [Window.Canvas] are not available then.
// This avoids unbounded memory growth in long-running load tests.
func Lean() OpenOption {
	return func(o *openOptions) {
		o.lean = true
	}
}

// sizeClassOf maps the width like the web frontend, see core.WindowSizeClass and eventhandling.ts.
func sizeClassOf(width int) proto.WindowSizeClass {
	switch {
	case width >= 1536:
		return proto.SizeClass2XL
	case width >= 1280:
		return proto.SizeClassXL
	case width >= 1024:
		return proto.SizeClassLarge
	case width >= 768:
		return proto.SizeClassMedium
	default:
		return proto.SizeClassSmall
	}
}

// transport connects a window with a scope, either within the same process or through a websocket.
type transport interface {
	// connect attaches the window to the scope and assigns the session.
	connect(w *Window) error
	// send delivers an event to the scope.
	send(w *Window, evt proto.NagoEvent) error
	// settle waits until the effects of all sent events have been received.
	settle(w *Window) error
	upload(w *Window, id string, files []core.File) error
	close(w *Window)
}

// A Window is a browser window connected to a nago application, either within the same process (see
// [App.Open]) or through a websocket (see [Dial]). It emulates the frontend: it keeps the latest rendered
// tree, follows navigation requests and records all other side effects. All actions must be called from a
// single goroutine.
type Window struct {
	t       TB
	tr      transport
	opts    openOptions
	scopeID proto.ScopeID
	rid     proto.RID
	closed  bool

	mutex       sync.Mutex
	changed     chan struct{} // closed and replaced whenever an event has been received
	answered    proto.RID     // highest request id answered by the backend
	sideEffects int           // amount of received events which prove that an action has been processed
	awaitRID    proto.RID     // request id of the last action which expects an answer
	awaitMark   int           // sideEffects when the last action has been sent
	connErr     error
	tree        proto.Component
	history     []Route
	events      []proto.NagoEvent
	calls       []*proto.CallRequested
	downloads   []core.ExportFilesOptions
	resources   []proto.Resource
	imports     []*proto.FileImportRequested
	opened      []Route
	links       []string
	focus       string
	clipboard   string
	errs        []string
	listeners   map[proto.Uint]listener
	canvasSeen  map[string]int
	navigation  []Route // pending navigations requested by the application
}

func newWindow(t TB, tr transport, path core.NavigationPath, opts []OpenOption) *Window {
	w := &Window{t: t, tr: tr, scopeID: proto.NewScopeID(), changed: make(chan struct{})}
	w.opts.locale = "de"
	w.opts.windowInfo = proto.WindowInfo{Density: 2, ColorScheme: proto.Light}
	Size(1280, 800)(&w.opts)
	for _, opt := range opts {
		opt(&w.opts)
	}

	if w.opts.sessionID == "" {
		w.opts.sessionID = string(proto.NewScopeID())
	}

	w.history = []Route{{Path: path, Values: w.opts.values}}
	return w
}

// open performs the handshake of the frontend and allocates the initial route.
func (w *Window) open() {
	w.t.Helper()

	start := time.Now()
	if err := w.tr.connect(w); err != nil {
		w.t.Fatalf("nagotest: cannot connect: %v", err)
	}

	w.t.Cleanup(w.Close)
	w.configure()
	w.allocate(w.Route())
	w.Settle()
	w.observe("open", string(w.Route().Path), start)
}

func (w *Window) configure() {
	w.t.Helper()

	w.dispatch(&proto.ScopeConfigurationChangeRequested{
		AcceptLanguage: proto.Locale(w.opts.locale),
		WindowInfo:     w.opts.windowInfo,
		RID:            w.nextRID(),
	})
}

func (w *Window) observe(kind, target string, start time.Time) {
	if w.opts.observer != nil {
		w.opts.observer(Action{Kind: kind, Target: target, Duration: time.Since(start)})
	}
}

func (w *Window) nextRID() proto.RID {
	w.rid++
	return w.rid
}

func (w *Window) dispatch(evt proto.NagoEvent) {
	w.t.Helper()

	if rid := ridOf(evt); rid != 0 {
		switch evt.(type) {
		case *proto.RootViewDestructionRequested:
			// never answered
		default:
			w.mutex.Lock()
			w.awaitRID = rid
			w.awaitMark = w.sideEffects
			w.mutex.Unlock()
		}
	}

	if err := w.tr.send(w, evt); err != nil {
		w.t.Fatalf("nagotest: cannot send %T: %v", evt, err)
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

// Settle waits until the window is idle and fails the test otherwise.
//
// Within the same process, this means that all events have been processed, pending state changes have been
// rendered, background tasks like [core.OnAppear] have completed and requested navigations have been
// followed. Delayed functions (see [core.Window.PostDelayed]) are not awaited.
//
// Through a websocket, it waits for the answer to the last action and follows requested navigations.
// If an action changes no state, the backend does not render, thus the window requests a render after
// [RenderTimeout]. Results of background tasks may arrive later, use [Window.WaitFor] for them.
//
// All actions of a Window settle automatically, thus you only need to call this after changing application
// state from outside.
func (w *Window) Settle() {
	w.t.Helper()

	for {
		if err := w.tr.settle(w); err != nil {
			w.t.Fatalf("nagotest: %v", err)
		}

		if !w.followNavigation() {
			break
		}
	}

	w.mutex.Lock()
	errs := w.errs
	w.errs = nil
	w.mutex.Unlock()

	for _, err := range errs {
		w.t.Errorf("nagotest: %s", err)
	}
}

// waitUntil blocks until cond, which is evaluated while holding the mutex, is true or the timeout
// elapsed. It returns false on timeout.
func (w *Window) waitUntil(timeout time.Duration, cond func() bool) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		w.mutex.Lock()
		ok := cond()
		changed := w.changed
		w.mutex.Unlock()

		if ok {
			return true
		}

		select {
		case <-changed:
		case <-timer.C:
			return false
		}
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

// WaitFor waits until at least one node matches m, e.g. for results of background tasks, and fails the test
// after the timeout.
func (w *Window) WaitFor(m Matcher, timeout time.Duration) Selection {
	w.t.Helper()

	start := time.Now()
	var sel Selection
	ok := w.waitUntil(timeout, func() bool {
		sel = Selection{w: w, matcher: m, nodes: find(w.tree, nil, m)}
		return len(sel.nodes) > 0
	})

	if !ok {
		w.t.Fatalf("nagotest: %s: not found within %v", m, timeout)
	}

	if _, remote := w.tr.(*wsTransport); remote {
		// A background task may have changed its state while the found tree was rendered, which causes another
		// render with new callbacks. Request a render as barrier, so that the next action refers to the latest tree.
		w.dispatch(&proto.RootViewRenderingRequested{RID: w.nextRID()})
	}

	w.Settle()
	w.observe("wait", m.desc, start)
	return w.FindAll(m)
}

// Close disconnects the window immediately, like closing the browser tab. Closing twice has no effect.
func (w *Window) Close() {
	if w.closed {
		return
	}

	w.closed = true
	w.tr.close(w)
}

// Reload emulates a browser reload: the root view is destroyed and allocated again, thus all window
// states are lost. The scope and the session are kept.
func (w *Window) Reload() {
	w.t.Helper()

	start := time.Now()
	w.dispatch(&proto.RootViewDestructionRequested{RID: w.nextRID()})
	w.allocate(w.Route())
	w.Settle()
	w.observe("reload", string(w.Route().Path), start)
}

// Reconnect emulates a lost connection: a new channel is connected to the existing scope, which is configured
// and requests the current root view again, as the frontend does. The window states are kept.
func (w *Window) Reconnect() {
	w.t.Helper()

	start := time.Now()
	w.tr.close(w)
	w.mutex.Lock()
	w.connErr = nil
	w.mutex.Unlock()
	if err := w.tr.connect(w); err != nil {
		w.t.Fatalf("nagotest: cannot reconnect: %v", err)
	}

	w.configure()
	w.allocate(w.Route())
	// the backend reuses the allocated root view without an answer, thus request the render explicitly
	w.dispatch(&proto.RootViewRenderingRequested{RID: w.nextRID()})
	w.Settle()
	w.observe("reconnect", string(w.Route().Path), start)
}

// Scope returns the underlying scope of a window within the same process and nil otherwise.
func (w *Window) Scope() *core.Scope {
	if tr, ok := w.tr.(*localTransport); ok {
		return tr.scope
	}

	return nil
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

// Downloads returns all files which have been sent using [core.Window.ExportFiles]. It is only available
// within the same process, see [Window.Resources] for a websocket.
func (w *Window) Downloads() []core.ExportFilesOptions {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return append([]core.ExportFilesOptions(nil), w.downloads...)
}

// Resources returns all resources which the application requested the browser to download. Through a
// websocket, this is the result of [core.Window.ExportFiles].
func (w *Window) Resources() []proto.Resource {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return append([]proto.Resource(nil), w.resources...)
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

// fail records a broken connection and wakes up all waiters.
func (w *Window) fail(err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.connErr == nil {
		w.connErr = err
	}

	close(w.changed)
	w.changed = make(chan struct{})
}

// receive is called for each event published by the scope.
func (w *Window) receive(evt proto.NagoEvent) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	close(w.changed)
	w.changed = make(chan struct{})

	if !w.opts.lean {
		w.events = append(w.events, evt)
	}

	if rid := ridOf(evt); rid > w.answered {
		w.answered = rid
	}

	switch evt := evt.(type) {
	case *proto.RootViewInvalidated:
		w.tree = evt.Root
	case *proto.NavigationForwardToRequested:
		w.sideEffects++
		route := routeOf(evt.RootView, evt.Values)
		if evt.Target != "" && evt.Target != "_self" {
			w.opened = append(w.opened, route)
			return
		}

		w.history = append(w.history, route)
		w.navigation = append(w.navigation, route)
	case *proto.NavigationBackRequested:
		w.sideEffects++
		if len(w.history) > 1 {
			w.history = w.history[:len(w.history)-1]
		}
		w.navigation = append(w.navigation, w.history[len(w.history)-1])
	case *proto.NavigationReplaceRequested:
		w.sideEffects++
		route := routeOf(evt.RootView, evt.Values)
		w.history[len(w.history)-1] = route
		w.navigation = append(w.navigation, route)
	case *proto.NavigationResetRequested:
		w.sideEffects++
		route := routeOf(evt.RootView, evt.Values)
		w.history = []Route{route}
		w.navigation = append(w.navigation, route)
	case *proto.NavigationReloadRequested:
		w.sideEffects++
		w.navigation = append(w.navigation, w.history[len(w.history)-1])
	case *proto.OpenHttpLink:
		w.sideEffects++
		w.links = append(w.links, string(evt.Url))
	case *proto.OpenHttpFlow:
		w.sideEffects++
		w.links = append(w.links, string(evt.Url))
	case *proto.FileImportRequested:
		w.sideEffects++
		w.imports = append(w.imports, evt)
	case *proto.SendMultipleRequested:
		w.sideEffects++
		w.resources = append(w.resources, evt.Resources...)
	case *proto.CallRequested:
		w.sideEffects++
		if !w.opts.lean {
			w.calls = append(w.calls, evt)
		}

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
		w.sideEffects++
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

func unexpected(evt any) error {
	return fmt.Errorf("%T is not a NagoEvent", evt)
}

// ridOf returns the request id of an event or 0.
func ridOf(evt proto.NagoEvent) proto.RID {
	if src, ok := evt.(interface{ GetRID() proto.RID }); ok {
		return src.GetRID()
	}

	return 0
}
