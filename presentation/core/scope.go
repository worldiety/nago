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
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/presentation/proto"
	"golang.org/x/text/language"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
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
	allocatedRootView std.Option[*scopeWindow]
	lifetime          time.Duration
	endOfLifeAt       atomic.Pointer[time.Time]
	channel           concurrent.Value[Channel]
	chanDestructor    concurrent.Value[func()]
	destroyed         concurrent.Value[bool]
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
	statesById         map[string]TransientProperty

	sessionID              session.ID
	sessionByID            session.FindUserSessionByID
	virtualSession         atomic.Pointer[session.UserSession]
	ignoreNextInvalidation atomic.Bool
	dirty                  bool
}

func NewScope(ctx context.Context, app *Application, tempRootDir string, id proto.ScopeID, lifetime time.Duration, factories map[proto.RootViewID]ComponentFactory, sessionByID session.FindUserSessionByID) *Scope {

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
		locale:      language.German, // TODO implement me
		statesById:  make(map[string]TransientProperty),
		sessionByID: sessionByID,
	}

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
		node := &proto.VStack{
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

func (s *Scope) ExportFilesOptions(id string) (ExportFilesOptions, bool) {
	s.Tick() // keep this scope alive
	if s.allocatedRootView.IsNone() {
		slog.Error("no such rootview allocated")
		return ExportFilesOptions{}, false
	}

	root := s.allocatedRootView.Unwrap()

	files, ok := root.exportFilesReceivers[id]
	if !ok {
		slog.Error("unknown import file", slog.Any("id", id))
		return ExportFilesOptions{}, false
	}

	return files, true
}

func (s *Scope) ImportFilesOptions(id string) (ImportFilesOptions, bool) {
	s.Tick() // keep this scope alive
	if s.allocatedRootView.IsNone() {
		slog.Error("no such rootview allocated")
		return ImportFilesOptions{}, false
	}

	root := s.allocatedRootView.Unwrap()

	files, ok := root.importFilesReceivers[id]
	if !ok {
		slog.Error("unknown import file", slog.Any("id", id))
		return ImportFilesOptions{}, false
	}

	return files, true
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

// Connect attaches the given channel to this Scope immediately. There must be exact 1 Scope per Channel.
// The use case is, that Scopes can be transferred from one channel to another easily.
// Note, that this is free of technical data races, however it may suffer from logical races, so do not connect
// concurrently, because things like destructor invocations and updates will logically race.
func (s *Scope) Connect(c Channel) {
	if c == nil {
		c = NopChannel{}
	}

	slog.Info("scope connected to channel", slog.String("scopeId", string(s.id)), slog.String("channel", fmt.Sprintf("%T", c)))

	if destructor := s.chanDestructor.Value(); destructor != nil {
		destructor()
	}
	s.channel.SetValue(c)

	s.chanDestructor.SetValue(c.Subscribe(func(msg []byte) error {
		defer s.eventLoop.Tick()
		return s.handleMessage(msg)
	}))
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

	if s.destroyed.Value() {
		slog.Error("scope is already destroyed but received a message", "sid", s.id, "what", fmt.Sprintf("%T", nagoEvt))
		return fmt.Errorf("scope already destroyed")
	}

	s.eventLoop.Post(func() {
		s.handleEvent(nagoEvt)

		var rid proto.RID
		if ridSrc, ok := nagoEvt.(interface{ GetRID() proto.RID }); ok {
			rid = ridSrc.GetRID()
		}

		if s.ignoreNextInvalidation.Load() {
			s.ignoreNextInvalidation.Store(false)
			return
		}

		if s.dirty || s.hasDirtyStates() {
			s.forceRender(rid)
			s.dirty = false
		}
	})

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

	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	tmp := proto.NewBinaryWriter(buf)
	if err := proto.Marshal(tmp, evt); err != nil {
		slog.Error("cannot marshal nago event", slog.Any("evt", evt))
		return
	}

	if err := s.channel.Value().Publish(buf.Bytes()); err != nil {
		slog.Error("cannot publish websocket message", "err", err, "scope", s.id, "destroyed", s.destroyed.Value())
	}
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
	}

	alloc := s.allocatedRootView.Unwrap()

	s.Publish(s.render(reqId, alloc))
}

// updateTick is called with a fixed rate. There is one application wide update ticker, thus this must not block at
// all.
func (s *Scope) updateTick(now time.Time) {
	s.eventLoop.Post(func() {
		// I can't estimate how expensive this becomes, to post for thousands of scopes at once. However, the updater
		// will throttle automatically, if it becomes to slow.
		if s.hasDirtyStates() {
			s.forceRender(0)
		}
	})
}

func (s *Scope) hasDirtyStates() bool {
	if s.allocatedRootView.IsNone() {
		return false
	}

	// TODO replace individual flags with a single flag per scope_window
	alloc := s.allocatedRootView.Unwrap()

	requiresRender := false
	for _, property := range alloc.states {
		if property.dirty() {
			requiresRender = true
			break
		}
	}

	if !requiresRender {
		for _, property := range s.statesById {
			if property.dirty() {
				requiresRender = true
				break
			}
		}
	}

	return requiresRender
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
				rn = &proto.VStack{
					Children: []proto.Component{
						&proto.TextView{Value: "panic during rendering, check server-side logs"},
					},
					Frame: proto.Frame{Width: "100%", Height: "100dvh"},
				}

			}
		}()
		return scopeWnd.render()
	}()

	return &proto.RootViewInvalidated{
		RID:  requestId,
		Root: renderResult,
	}
}

// Destroy frees all allocated components and removes factory pointers.
// The scope is of no use afterward.
// Do never call this from the event loop.
// Note, that this may race logically when called concurrently.
func (s *Scope) Destroy() {
	fmt.Println("scope.Destroy")
	if !concurrent.CompareAndSwap(&s.destroyed, false, true) {
		return
	}

	s.eventLoop.Post(func() {
		// the event loop is panic protected, thus separate the observer execution
		for _, f := range s.onDestroyObservers.PopAll() {
			f()
		}
	})
	s.eventLoop.Post(func() {
		s.destroy()
	})

	s.eventLoop.Destroy()

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
}

// only for event loop
func (s *Scope) handleSessionAssigned(evt *proto.SessionAssigned) {
	s.sessionID = session.ID(evt.SessionID)
	tmp := s.sessionByID(s.sessionID)
	s.virtualSession.Store(&tmp)
}
