package core

import (
	"context"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/presentation/ora"
	"golang.org/x/text/language"
	"log/slog"
	"os"
	"path/filepath"
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
	app                *Application
	id                 ora.ScopeID
	factories          map[ora.ComponentFactoryId]ComponentFactory
	allocatedRootView  std.Option[*scopeWindow]
	lifetime           time.Duration
	endOfLifeAt        atomic.Pointer[time.Time]
	channel            concurrent.Value[Channel]
	chanDestructor     concurrent.Value[func()]
	destroyed          concurrent.Value[bool]
	eventLoop          *EventLoop
	lastMessageType    ora.EventType
	ctx                context.Context
	cancelCtx          func()
	sessionID          SessionID
	tempDirMutex       sync.Mutex
	tempRootDir        string
	tempDir            string
	nextFileSeqNo      int64
	windowInfo         WindowInfo
	onDestroyObservers concurrent.Slice[func()]
	location           *time.Location
	subject            concurrent.Value[auth.Subject]
	locale             language.Tag
}

func NewScope(ctx context.Context, app *Application, tempRootDir string, id ora.ScopeID, lifetime time.Duration, factories map[ora.ComponentFactoryId]ComponentFactory) *Scope {

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
	}

	loc, err := time.LoadLocation("Europe/Berlin") // TODO implement me
	if err != nil {
		slog.Error("cannot load location", slog.Any("err", err))
		loc = time.UTC
	}
	s.location = loc

	s.subject.SetValue(auth.InvalidSubject{})

	s.eventLoop.SetOnPanicHandler(func(p any) {
		s.Publish(ora.ErrorOccurred{
			Type:    ora.ErrorOccurredT,
			Message: fmt.Sprintf("panic in event loop: %v", p),
		})
	})
	s.channel.SetValue(NopChannel{})
	s.Tick()

	return s
}

func (s *Scope) ID() ora.ScopeID {
	return s.id
}

func (s *Scope) ExportFilesOptions(id string) (ExportFilesOptions, bool) {
	s.Tick() // keep this scope alive
	root, err := s.allocatedRootView.Get()
	if err != nil {
		slog.Error("no such rootview allocated")
		return ExportFilesOptions{}, false
	}

	files, ok := root.exportFilesReceivers[id]
	if !ok {
		slog.Error("unknown import file", slog.Any("id", id))
		return ExportFilesOptions{}, false
	}

	return files, true
}

func (s *Scope) ImportFilesOptions(id string) (ImportFilesOptions, bool) {
	s.Tick() // keep this scope alive
	root, err := s.allocatedRootView.Get()
	if err != nil {
		slog.Error("no such rootview allocated")
		return ImportFilesOptions{}, false
	}

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
	if s.allocatedRootView.Valid {
		if s.allocatedRootView.Valid {
			s.allocatedRootView.Unwrap().Invalidate()
		}
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

	t, err := ora.Unmarshal(buf)
	if err != nil {
		return err
	}

	s.eventLoop.Post(func() {
		s.handleEvent(t, true)

		// todo the vue frontend sends a lot of empty transactions on mouse movements and vuejs makes garbage out of the viewtree
		wasEmptyTx := false
		if tx, ok := t.(ora.EventsAggregated); ok {
			wasEmptyTx = len(tx.Events) == 0
		}

		wasDestructed := isEvent[ora.ComponentDestructionRequested](t) || isEvent[ora.ScopeDestructionRequested](t)

		// todo handleEvent may have caused already a rendering. Should we omit to avoid sending multiple times?
		if !wasDestructed && !wasEmptyTx && s.lastMessageType != ora.ComponentInvalidatedT {
			s.forceRender()
		}
	})

	return nil
}

func isEvent[T ora.Event](e ora.Event) bool {
	if _, ok := e.(T); ok {
		return ok
	}

	if tx, ok := e.(ora.EventsAggregated); ok {
		for _, event := range tx.Events {
			if _, ok := event.(T); ok {
				return ok
			}
		}
	}

	return false
}

func (s *Scope) Publish(evt ora.Event) {
	switch evt := evt.(type) {
	case ora.ComponentInvalidated:
		s.lastMessageType = evt.Type
	default:
		s.lastMessageType = ""
	}

	if err := s.channel.Value().Publish(ora.Marshal(evt)); err != nil {
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

// sendAck eventually sends an acknowledged message, if id is not 0. This is intentional, so a sender
// can optimize for performance.
func (s *Scope) sendAck(id ora.RequestId) {
	if id == 0 {
		return
	}

	s.Publish(ora.Acknowledged{
		Type:      ora.AcknowledgedT,
		RequestId: id,
	})
}

func (s *Scope) sendPing() {
	s.Publish(ora.Ping{
		Type: ora.PingT,
	})
}

// only for event loop
func (s *Scope) forceRender() {
	alloc, err := s.allocatedRootView.Get()
	if err != nil {
		slog.Error("no view to render is allocated", "err", err)
		return
	}

	s.Publish(s.render(0, alloc))
}

// updateTick is called with a fixed rate. There is one application wide update ticker, thus this must not block at
// all.
func (s *Scope) updateTick(now time.Time) {
	s.eventLoop.Post(func() {
		// I can't estimate how expensive this becomes, to post for thousands of scopes at once. However, the updater
		// will throttle automatically, if it becomes to slow.
		alloc, err := s.allocatedRootView.Get()
		if err != nil {
			return
		}

		requiresRender := false
		for _, property := range alloc.states {
			if property.dirty() {
				requiresRender = true
				break
			}
		}

		if requiresRender {
			s.forceRender()
		}
	})
}

// only for event loop
func (s *Scope) render(requestId ora.RequestId, scopeWnd *scopeWindow) ora.ComponentInvalidated {

	return ora.ComponentInvalidated{
		Type:      ora.ComponentInvalidatedT,
		RequestId: requestId,
		Component: scopeWnd.render(),
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

	alloc, err := s.allocatedRootView.Get()
	if err != nil {
		return
	}

	alloc.destroy()

}

// only for event loop
func (s *Scope) handleSessionAssigned(evt ora.SessionAssigned) {
	s.sessionID = SessionID(evt.SessionID)
}
