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
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/presentation/proto"
	"golang.org/x/text/language"
)

type Destroyable interface {
	Destroy()
}

type ComponentFactory func(Window) View

// A Scope manage its own area of associated pointers. A Pointer must be only unique per Scope.
// Resolving or keeping pointers outside a scope is inherently unsafe (e.g. a lookup map).
// A Scope can hold an arbitrary amount of components, created by an arbitrary amount of factories.
// This is intentional. E.g. a mobile app can create multiple instances of the same (or different)
// "pages" and those pages may communicate immediately with each other (observer etc.) without
// causing race conditions.
// Each Scope has a single event loop to guarantee race free event processing.
// A Scope is probably a single window.
// To allow interacting between different scopes, we may either go the full serializing message route or
// as a cheap alternative, replace all event loops with the same looper instance.
// However, we must be careful on destruction of the scopes sharing them.
type Scope struct {
	app               *Application
	id                proto.ScopeID
	factories         map[proto.RootViewID]ComponentFactory
	allocatedRootView option.Opt[*scopeWindow]
	lifetime          time.Duration
	endOfLifeAt       atomic.Pointer[time.Time]
	channelMutex      sync.Mutex
	channel           concurrent.Value[Channel]
	chanDestructor    concurrent.Value[func()]
	destroyed         atomic.Bool
	eventLoop         *EventLoop
	ctx               context.Context
	cancelCtx         func()

	tempDirMutex       sync.Mutex
	tempRootDir        string
	tempDir            string
	nextFileSeqNo      int64
	windowInfo         WindowInfo
	onDestroyObservers concurrent.Slice[func()]
	location           *time.Location
	subject            concurrent.Value[auth.Subject]
	locale             language.Tag
	bundle             *i18n.Bundle
	// statesById holds the transient states, see [TransientStateOf]. They may be allocated from any goroutine,
	// thus the map is protected by statesMutex.
	statesById  map[string]TransientProperty
	statesMutex sync.Mutex
	// transientGeneration is the generation of the last render which marked the transient states as rendered.
	transientGeneration int64
	// ids hands out the pointers and ids of this scope, see identifiers.
	ids identifiers
	// generation is the render generation of this scope. It is monotonic across all windows of this scope,
	// because the transient states outlive a window.
	generation atomic.Int64

	sessionID      session.ID
	sessionByID    session.FindUserSessionByID
	virtualSession atomic.Pointer[session.UserSession]
	// sessionFromTransport is set, once the transport assigned the session, e.g. from an http-only cookie. A
	// client must not choose its session then, see [Scope.AssignSession].
	sessionFromTransport   atomic.Bool
	ignoreNextInvalidation atomic.Bool
	dirty                  bool
	// background counts running goroutines started on behalf of this scope, e.g. by [OnAppear].
	background atomic.Int64
	// the uploads and downloads of the current window, which are looked up by HTTP handlers
	filesMutex  sync.Mutex
	importFiles map[string]ImportFilesOptions
	exportFiles map[string]ExportFilesOptions

	// tickQueued is true, while a function of the update ticker is posted but not yet executed.
	tickQueued atomic.Bool
	// staleCalls counts the discarded requests of stale trees, see discardStale.
	staleCalls atomic.Int64
	staleSince time.Time
	staleCount int
}

func NewScope(ctx context.Context, app *Application, tempRootDir string, id proto.ScopeID, lifetime time.Duration, factories map[proto.RootViewID]ComponentFactory, sessionByID session.FindUserSessionByID) *Scope {

	defaultLang := language.English
	defaultBundle, ok := i18n.Default.MatchBundle(defaultLang)
	if !ok {
		slog.Error("no bundle for default language")
		i18n.StringKey("no bundle for default language")
		i18n.Default.Flush()
		defaultBundle, ok = i18n.Default.MatchBundle(defaultLang)
		if !ok {
			panic(fmt.Errorf("implementation error of i18n package: configured at least a single string but broke"))
		}
	}

	scopeCtx, cancel := context.WithCancel(ctx)
	s := &Scope{
		app:         app,
		id:          id,
		factories:   factories,
		lifetime:    lifetime,
		eventLoop:   NewEventLoop(),
		ctx:         scopeCtx,
		cancelCtx:   cancel,
		tempRootDir: tempRootDir,
		locale:      defaultLang,
		bundle:      defaultBundle,
		statesById:  make(map[string]TransientProperty),
		sessionByID: sessionByID,
	}
	s.ids.init()

	loc, err := time.LoadLocation("Europe/Berlin") // TODO implement me
	if err != nil {
		slog.Error("cannot load location", slog.Any("err", err))
		loc = time.UTC
	}
	s.location = loc

	s.subject.SetValue(s.app.getAnonUser())

	s.eventLoop.SetOnPanicHandler(func(p any) {
		/*s.Publish(proto.ErrorOccurred{
			Type:    proto.ErrorOccurredT,
			Message: fmt.Sprintf("panic in event loop: %v", p),
		})*/
		node := &proto.Stack{
			Orientation: proto.Vertical,
			Children: []proto.Component{
				&proto.TextView{Value: "panic during event loop, check server-side logs"},
			},
			Frame: proto.Frame{Width: "100%", Height: "100dvh"},
		}

		s.Publish(&proto.RootViewInvalidated{
			RID:  0,
			Root: node,
		})
	})
	s.channel.SetValue(NopChannel{})
	s.Tick()

	return s
}

func (s *Scope) ID() proto.ScopeID {
	return s.id
}

// ExportFilesOptions returns the download of the current window with the given id. It may be called from any
// goroutine, e.g. by an HTTP handler.
func (s *Scope) ExportFilesOptions(id string) (ExportFilesOptions, bool) {
	s.Tick() // keep this scope alive

	s.filesMutex.Lock()
	files, ok := s.exportFiles[id]
	s.filesMutex.Unlock()

	if !ok {
		slog.Error("unknown export file", slog.Any("id", id))
		return ExportFilesOptions{}, false
	}

	return files, true
}

// ImportFilesOptions returns the upload of the current window with the given id. It may be called from any
// goroutine, e.g. by an HTTP handler.
func (s *Scope) ImportFilesOptions(id string) (ImportFilesOptions, bool) {
	s.Tick() // keep this scope alive

	s.filesMutex.Lock()
	files, ok := s.importFiles[id]
	s.filesMutex.Unlock()

	if !ok {
		slog.Error("unknown import file", slog.Any("id", id))
		return ImportFilesOptions{}, false
	}

	return files, true
}

func (s *Scope) putImportFiles(options ImportFilesOptions) {
	s.filesMutex.Lock()
	defer s.filesMutex.Unlock()

	if s.importFiles == nil {
		s.importFiles = map[string]ImportFilesOptions{}
	}

	s.importFiles[options.ID] = options
}

func (s *Scope) putExportFiles(options ExportFilesOptions) {
	s.filesMutex.Lock()
	defer s.filesMutex.Unlock()

	if s.exportFiles == nil {
		s.exportFiles = map[string]ExportFilesOptions{}
	}

	s.exportFiles[options.ID] = options
}

// clearFiles forgets the uploads and downloads, when their window goes away.
func (s *Scope) clearFiles() {
	s.filesMutex.Lock()
	defer s.filesMutex.Unlock()

	s.importFiles = nil
	s.exportFiles = nil
}

func (s *Scope) getTempDir() (string, error) {
	s.tempDirMutex.Lock()
	defer s.tempDirMutex.Unlock()

	if s.tempDir != "" {
		return s.tempDir, nil
	}

	// we don't know where the temp root is. It may be in our apps home (e.g. in shared hosting environments)
	// or in the systems temp dir.
	path := filepath.Join(s.tempRootDir)
	//0700 means that only the owner can read and write the dir, files are 0600
	if err := os.MkdirAll(path, 0700); err != nil {
		return "", fmt.Errorf("cannot create temp dir for scope: %s: %w", path, err)
	}

	slog.Info("created temp dir for scope", "path", path)

	s.tempDir = path
	return path, nil
}

func (s *Scope) updateWindowInfo(winfo WindowInfo) {
	s.windowInfo = winfo
	if s.allocatedRootView.IsSome() {
		s.allocatedRootView.Unwrap().Invalidate()
	}
}

func (s *Scope) updateLanguage(locale string) {
	tags, _, err := language.ParseAcceptLanguage(locale)
	if err != nil {
		slog.Error("cannot parse language", slog.Any("err", err))
		return
	}

	if len(tags) == 0 {
		slog.Error("cannot parse language, no tags", "text", locale)
		return
	}

	// todo the race condition situation is not clear here, normally this is protected by the event looper, but subject, scope and window may leak into other go routines
	s.locale = tags[0]
	if b, ok := i18n.Default.MatchBundle(s.locale); ok {
		s.bundle = b

		subject := s.subject.Value()
		if setter, ok := subject.(subjectLanguageSetter); ok {
			setter.SetLanguage(s.locale)
			setter.SetBundle(s.bundle)
		}
	} else {
		slog.Error("cannot match bundle language", "tag", s.locale)
	}

}

// Connect attaches the given channel to this Scope immediately. There must be exact 1 Scope per Channel.
// The use case is, that Scopes can be transferred from one channel to another easily.
// Note, that this is free of technical data races, however it may suffer from logical races, so do not connect
// concurrently, because things like destructor invocations and updates will logically race.
func (s *Scope) Connect(c Channel) {
	s.channelMutex.Lock()
	defer s.channelMutex.Unlock()

	if c == nil {
		c = NopChannel{}
	}

	//slog.Info("scope connected to channel", slog.String("scopeId", string(s.id)), slog.String("channel", fmt.Sprintf("%T", c)))

	if destructor := s.chanDestructor.Value(); destructor != nil {
		destructor()
	}
	s.channel.SetValue(c)

	s.chanDestructor.SetValue(c.Subscribe(func(msg []byte) error {
		return s.handleMessage(msg)
	}))
}

// Disconnect detaches the given channel, but only if it is still connected. A channel which has been replaced
// by a newer connection, e.g. due to a reconnect, does not affect the newer one.
func (s *Scope) Disconnect(c Channel) {
	s.channelMutex.Lock()
	connected := s.channel.Value() == c
	s.channelMutex.Unlock()

	if connected {
		s.Connect(nil)
	}
}

func (s *Scope) handleMessage(buf []byte) error {
	s.Tick()

	t, err := proto.Unmarshal(proto.NewBinaryReader(bytes.NewBuffer(buf)))
	if err != nil {
		return err
	}

	nagoEvt, ok := t.(proto.NagoEvent)
	if !ok {
		return fmt.Errorf("protocol error while handle message: %T is not a proto.NagoEvent", t)
	}

	return s.Dispatch(nagoEvt)
}

// Dispatch processes the given event as if it has been received from the connected [Channel]. The event is
// handled asynchronously by the event loop of this scope. This is intended for transports within the same
// process, see also [EventChannel].
func (s *Scope) Dispatch(nagoEvt proto.NagoEvent) error {
	s.Tick()

	posted := s.eventLoop.Post(func() {
		s.handleEvent(nagoEvt)

		var rid proto.RID
		if ridSrc, ok := nagoEvt.(interface{ GetRID() proto.RID }); ok {
			rid = ridSrc.GetRID()
		}

		if s.ignoreNextInvalidation.Load() {
			s.ignoreNextInvalidation.Store(false)
			return
		}

		// the client can create logical races by sending multiple messages right after each other, which
		// may cause the first render to succeed but get skipped by the client because it waits already
		// for the second render, which never comes because the first render already
		// marked all states as clean and the second does not
		// mutate any state. To avoid this, the backend MUST mutate a state to ensure that at least one render gets
		// through. If we skip this optimization, any client message would cause a rendering, which is
		// not desirable either, especially for high-performance canvas operations.
		if s.dirty || s.hasDirtyStates() {
			s.forceRender(rid)
			s.dirty = false
		}
	})

	if !posted {
		slog.Error("scope is already destroyed but received a message", "sid", logging.Secret(string(s.id)), "what", fmt.Sprintf("%T", nagoEvt))
		return fmt.Errorf("scope already destroyed")
	}

	return nil
}

func (s *Scope) Publish(evt proto.NagoEvent) {
	//switch evt := evt.(type) {

	// TODO fix me and think again
	/*case proto.ComponentInvalidated:
		if s.ignoreNextInvalidation.Load() {
			s.ignoreNextInvalidation.Store(false)
			return
		}

		s.lastMessageType = evt.Type
	case proto.Acknowledged:
		// ignore
	default:
		s.lastMessageType = ""*/
	//}

	channel := s.channel.Value()
	if evtChan, ok := channel.(EventChannel); ok {
		if err := evtChan.PublishEvent(evt); err != nil {
			slog.Error("cannot publish event", "err", err, "scope", logging.Secret(string(s.id)), "destroyed", s.destroyed.Load())
		}

		return
	}

	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	tmp := proto.NewBinaryWriter(buf)
	if err := proto.Marshal(tmp, evt); err != nil {
		slog.Error("cannot marshal nago event", slog.Any("evt", evt))
		return
	}

	if err := channel.Publish(buf.Bytes()); err != nil {
		slog.Error("cannot publish websocket message", "err", err, "scope", logging.Secret(string(s.id)), "destroyed", s.destroyed.Load())
	}
}

// Flush blocks until all functions posted to the event loop so far have been processed. If any state is dirty
// afterward, a render is published. It returns true, if the scope is idle, which means that nothing else
// has been posted in the meantime, no background work (see [OnAppear]) is running and no state has been changed
// since the render. A destroyed scope is always idle, after its event loop has executed the remaining
// functions and exited. Note that delayed functions (see [Window.PostDelayed]), [OnFrame] and goroutines which
// are not started by nago are not considered.
// A state which is changed by each render, keeps the scope busy forever, see also [State.Set].
// Flush must never be called from the event loop, because that would deadlock.
func (s *Scope) Flush() (idle bool) {
	idle, _ = s.FlushTimeout(0)
	return idle
}

// FlushTimeout is like [Scope.Flush] but waits at most for the given timeout, e.g. because a function blocks
// the event loop. A timeout <= 0 waits forever. It returns false, if the timeout elapsed.
func (s *Scope) FlushTimeout(timeout time.Duration) (idle, ok bool) {
	var expired <-chan time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		expired = timer.C
	}

	done := make(chan bool, 1)
	posted := s.eventLoop.Post(func() {
		if s.dirty || s.hasDirtyStates() {
			s.forceRender(0)
			s.dirty = false
		}

		// The order of the following checks is important. A background goroutine applies its state changes and
		// posts before it is counted as done. Thus, if no background work is left, all of its effects are
		// visible to the checks below, including those of goroutines started by the render above.
		background := s.background.Load()
		dirty := s.dirty || s.hasDirtyStates()
		// the currently executed function is still pending
		done <- background == 0 && !dirty && s.eventLoop.Pending() <= 1
	})

	if !posted {
		select {
		case <-s.eventLoop.Done():
			return true, true
		case <-expired:
			return false, false
		}
	}

	select {
	case <-s.eventLoop.Done():
		return true, true
	case idle = <-done:
		return idle, true
	case <-expired:
		return false, false
	}
}

// goBackground runs fn in a new goroutine and accounts it as background work of this scope, see [Scope.Flush].
func (s *Scope) goBackground(fn func()) {
	s.background.Add(1)
	go func() {
		defer s.background.Add(-1)
		fn()
	}()
}

// Tick marks this scope as used and moves the EOL forward.
func (s *Scope) Tick() {
	eol := time.Now().Add(s.lifetime)
	s.endOfLifeAt.Store(&eol)
}

// EOL returns the current estimated end of life.
func (s *Scope) EOL() time.Time {
	return *s.endOfLifeAt.Load()
}

// only for event loop
func (s *Scope) forceRender(reqId proto.RID) {
	if s.allocatedRootView.IsNone() {
		s.Publish(&proto.ErrorRootViewAllocationRequired{RID: reqId})
		return
	}

	alloc := s.allocatedRootView.Unwrap()

	s.Publish(s.render(reqId, alloc))
}

// updateTick is called with a fixed rate. There is one application wide update ticker, thus this must not block at
// all. At most one tick is queued per scope, so that a busy or hanging event loop does not accumulate ticks.
func (s *Scope) updateTick(now time.Time) {
	if !s.tickQueued.CompareAndSwap(false, true) {
		return
	}

	posted := s.eventLoop.Post(func() {
		s.tickQueued.Store(false)
		if s.hasDirtyStates() {
			s.forceRender(0)
		} else if s.allocatedRootView.IsSome() {
			// an idle window must not keep the keys of its former trees
			s.allocatedRootView.Unwrap().pruneFormer(now)
		}
	})

	if !posted {
		s.tickQueued.Store(false)
	}
}

func (s *Scope) hasDirtyStates() bool {
	if s.allocatedRootView.IsNone() {
		return false
	}

	alloc := s.allocatedRootView.Unwrap()

	// window scoped states are aggregated in O(1) via a single dirty marker per window.
	if alloc.hasDirtyStates() {
		return true
	}

	// transient (scope/session level) states are not attached to a window and are therefore still checked
	// individually. Their amount is expected to be small.
	s.statesMutex.Lock()
	defer s.statesMutex.Unlock()

	for _, property := range s.statesById {
		if property.dirty() {
			return true
		}
	}

	return false
}

// setTransientGeneration marks all transient states as rendered by the given generation, see [scopeWindow.render].
func (s *Scope) setTransientGeneration(generation int64) {
	s.statesMutex.Lock()
	defer s.statesMutex.Unlock()

	s.transientGeneration = generation
	for _, property := range s.statesById {
		property.setGeneration(generation)
	}
}

// only for event loop
func (s *Scope) render(requestId proto.RID, scopeWnd *scopeWindow) *proto.RootViewInvalidated {

	renderResult := func() (rn RenderNode) {
		defer func() {
			if r := recover(); r != nil {
				if s.app.IsDebug() {
					fmt.Println(r)
					debug.PrintStack()
				} else {
					slog.Error(fmt.Sprintf("%v", r), slog.String("panic", string(debug.Stack())))
				}
				rn = &proto.Stack{
					Orientation: proto.Vertical,
					Children: []proto.Component{
						&proto.TextView{Value: "panic during rendering, check server-side logs"},
					},
					Frame: proto.Frame{Width: "100%", Height: "100dvh"},
				}

				// the client never receives the callbacks of the failed tree, so nothing may call them
				scopeWnd.dropCallbacks()
			}
		}()
		return scopeWnd.render()
	}()

	scopeWnd.checkRenderLoop(s.hasDirtyStates())

	return &proto.RootViewInvalidated{
		RID:  requestId,
		Root: renderResult,
	}
}

// Destroy frees all allocated components and removes factory pointers.
// The scope is of no use afterward. Functions which have already been posted to the event loop are still
// executed before, but any later post is rejected. Destroy does not block and may also be called from the
// event loop. See also [Application.DestroyScope] which also removes the scope from the application.
func (s *Scope) Destroy() {
	if !s.destroyed.CompareAndSwap(false, true) {
		return
	}

	s.eventLoop.Destroy(
		func() {
			// the event loop is panic protected, thus separate the observer execution
			for _, f := range s.onDestroyObservers.PopAll() {
				f()
			}
		},
		func() {
			s.destroy()
		},
	)
}

func (s *Scope) AddOnDestroyObserver(f func()) {
	s.onDestroyObservers.Append(f)
}

// only for event loop
func (s *Scope) destroy() {
	//if s.destroyed.Value() {
	//	return
	//}
	//
	//s.destroyed.SetValue(true)

	s.cancelCtx()

	for _, f := range s.onDestroyObservers.PopAll() {
		f()
	}

	s.channel.SetValue(NewPrintChannel()) // detach

	s.onDestroyObservers.Clear()
	//	clear(s.factories) // clearing this map would cause a data race, even though we use the factory as read-only

	if s.allocatedRootView.IsNone() {
		return
	}

	alloc := s.allocatedRootView.Unwrap()
	alloc.destroy()
	s.clearFiles()
}

// AssignSession binds the scope to the session, which the transport authenticated, e.g. by the http-only cookie
// of a websocket handshake. From then on, a [proto.SessionAssigned] sent by the client is ignored, because a
// client could otherwise choose an arbitrary session, bypassing the cookie. Clients without such a transport,
// like native apps, keep assigning their session by the event.
func (s *Scope) AssignSession(id session.ID) {
	s.Tick()
	s.eventLoop.Post(func() {
		s.sessionFromTransport.Store(true)
		s.assignSession(id)
	})
}

// only for event loop
func (s *Scope) handleSessionAssigned(evt *proto.SessionAssigned) {
	if s.sessionFromTransport.Load() {
		slog.Warn("ignored a session assigned by the client, the transport already assigned it", "scope", s.id)
		return
	}

	s.assignSession(session.ID(evt.SessionID))
}

// sessionLoggedOut renders the window of the scope as anonymous, if it belongs to the session and still shows a
// valid subject.
func (s *Scope) sessionLoggedOut(id session.ID) {
	s.eventLoop.Post(func() {
		if s.sessionID != id || s.allocatedRootView.IsNone() || !s.subject.Value().Valid() {
			return
		}

		s.allocatedRootView.Unwrap().UpdateSubject(nil)
		s.forceRender(0)
	})
}

// only for event loop
func (s *Scope) assignSession(id session.ID) {
	s.sessionID = id
	tmp := s.sessionByID(s.sessionID)
	_ = tmp.User() //issue a refresh, if required
	s.virtualSession.Store(&tmp)
}

type subjectLanguageSetter interface {
	SetLanguage(tag language.Tag)
	SetBundle(bundle *i18n.Bundle)
}
