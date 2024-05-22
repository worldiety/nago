package core

import (
	"context"
	"fmt"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/ora"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

type Destroyable interface {
	Destroy()
}

type allocatedComponent struct {
	Window      *scopeWindow
	Component   Component
	RenderState *RenderState
}

type ComponentFactory func(Window, ora.NewComponentRequested) Component

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
	app                 *Application
	id                  ora.ScopeID
	factories           map[ora.ComponentFactoryId]ComponentFactory
	allocatedComponents map[ora.Ptr]allocatedComponent
	lifetime            time.Duration
	endOfLifeAt         atomic.Pointer[time.Time]
	channel             AtomicRef[Channel]
	chanDestructor      AtomicRef[func()]
	destroyed           AtomicRef[bool]
	eventLoop           *EventLoop
	lastMessageType     ora.EventType
	ctx                 context.Context
	cancelCtx           func()
	sessionID           SessionID
	tempDirMutex        sync.Mutex
	tempRootDir         string
	tempDir             string
	nextFileSeqNo       int64
	onDestroyObservers  []func()
}

func NewScope(ctx context.Context, app *Application, tempRootDir string, id ora.ScopeID, lifetime time.Duration, factories map[ora.ComponentFactoryId]ComponentFactory) *Scope {

	scopeCtx, cancel := context.WithCancel(ctx)
	s := &Scope{
		app:                 app,
		id:                  id,
		factories:           factories,
		allocatedComponents: map[ora.Ptr]allocatedComponent{},
		lifetime:            lifetime,
		eventLoop:           NewEventLoop(),
		ctx:                 scopeCtx,
		cancelCtx:           cancel,
		tempRootDir:         tempRootDir,
	}

	s.eventLoop.SetOnPanicHandler(func(p any) {
		s.Publish(ora.ErrorOccurred{
			Type:    ora.ErrorOccurredT,
			Message: fmt.Sprintf("panic in event loop: %v", p),
		})
	})
	s.channel.Store(NopChannel{})
	s.Tick()

	return s
}

// OnFilesReceived provides a side channel for sending large streams of blobs which must not
// block the lightweight and responsive connected event Channel.
// If e.g. this scope is part of a http web server, each multipart form (one or many files) must trigger a call here.
// Note, that the origin could also be from different sources, like the content resolver within an Android App
// directly issued over FFI calls or other activity intents.
// The so-stored files are kept an undefined amount of time but at least as long the callback runs.
// However, the system may reclaim the used disk space if running short on storage space or it may keep it for
// even for years.
// So, to ensure a correct cleanup, use [FS.Clear] to remove all temporary files.
func (s *Scope) OnFilesReceived(receiverPtr ora.Ptr, it iter.Seq2[File, error]) error {
	s.Tick() // keep this scope alive

	s.eventLoop.Post(func() {
		var receiver Component
		if receiverPtr.Nil() {
			slog.Info("scope received unrequested files, trying to dispatch to any allocated root component...")

			dispatched := false
			for _, holder := range s.allocatedComponents {
				if rec, ok := holder.Component.(FilesReceiver); ok {
					if err := rec.OnFilesReceived(it); err != nil {
						dispatched = false
						slog.Error("failed to dispatch files", "err", err)
						break
					} else {
						dispatched = true
					}
				}
			}

			if !dispatched {
				slog.Error("scope received unrequested files, but could not find any allocated root component, file are lost")
				if err := Release(it); err != nil {
					slog.Error("cannot release received but unprocessed files", "err", err)
				}
			}

			return
		} else {
			if cmp, ok := s.allocatedComponents[receiverPtr]; ok {
				receiver = cmp.Component
			} else {
				for _, holder := range s.allocatedComponents {
					if cmp, ok := holder.RenderState.elements[receiverPtr]; ok {
						receiver = cmp
					}
				}
			}
		}

		if receiver == nil {
			slog.Error("receiver component for data stream not found")
			return
		}

		switch receiver := receiver.(type) {
		case FilesReceiver:
			if err := receiver.OnFilesReceived(it); err != nil {
				slog.Error("failed to dispatch files", "err", err)
				if err := Release(it); err != nil {
					slog.Error("cannot release received but unprocessed files", "err", err)
				}
			}
		default:
			slog.Error("receiver component for data stream has no compatible receiver interface, files are lost", "type", fmt.Sprintf("%T", receiver))
			if err := Release(it); err != nil {
				slog.Error("cannot release received but unprocessed files", "err", err)
			}
		}

		s.forceRender() // the callback likely changed some domain state, so invalidate
		s.eventLoop.Tick()
	})

	s.eventLoop.Tick() // trigger event loop processing, so that our post is actually processed.

	return nil
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

// Connect attaches the given channel to this Scope immediately. There must be exact 1 Scope per Channel.
// The use case is, that Scopes can be transferred from one channel to another easily.
func (s *Scope) Connect(c Channel) {
	if c == nil {
		c = NopChannel{}
	}

	slog.Info("scope connected to channel", slog.String("scopeId", string(s.id)), slog.String("channel", fmt.Sprintf("%T", c)))

	s.chanDestructor.With(func(destructor func()) func() {
		s.channel.Store(c)

		if destructor != nil {
			destructor()
		}

		return c.Subscribe(func(msg []byte) error {
			defer s.eventLoop.Tick()
			return s.handleMessage(msg)
		})
	})
}

func (s *Scope) handleMessage(buf []byte) error {
	s.Tick()

	t, err := ora.Unmarshal(buf)
	if err != nil {
		return err
	}

	s.eventLoop.Post(func() {
		s.handleEvent(t, true)
		// todo handleEvent may have caused already a rendering. Should we omit to avoid sending multiple times?
		if s.lastMessageType != ora.ComponentInvalidatedT {
			s.renderIfRequired()
		}
	})

	return nil
}

func (s *Scope) Publish(evt ora.Event) {
	switch evt := evt.(type) {
	case ora.ComponentInvalidated:
		s.lastMessageType = evt.Type
	default:
		s.lastMessageType = ""
	}

	if err := s.channel.Load().Publish(ora.Marshal(evt)); err != nil {
		slog.Error(err.Error())
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
func (s *Scope) renderIfRequired() {
	for _, component := range s.allocatedComponents {
		if IsDirty(component.Component) {
			//slog.Info("component is dirty", slog.Int("ptr", int(component.Component.ID())))
			s.Publish(s.render(0, component.Component))
		}
	}
}

// only for event loop
func (s *Scope) forceRender() {
	for _, component := range s.allocatedComponents {
		//slog.Info("component is dirty", slog.Int("ptr", int(component.Component.ID())))
		s.Publish(s.render(0, component.Component))
	}
}

// only for event loop
func (s *Scope) render(requestId ora.RequestId, component Component) ora.ComponentInvalidated {
	Freeze(component)
	defer Unfreeze(component)

	rs := s.allocatedComponents[component.ID()].RenderState
	rs.Clear()
	rs.Scan(component)
	ClearDirty(component)

	return ora.ComponentInvalidated{
		Type:      ora.ComponentInvalidatedT,
		RequestId: requestId,
		Component: component.Render(),
	}
}

// Destroy frees all allocated components and removes factory pointers.
// The scope is of no use afterward.
// Do never call this from the event loop.
func (s *Scope) Destroy() {
	s.destroyed.With(func(destroyed bool) bool {
		if destroyed {
			return true
		}

		var tmp []func()
		tmp = slices.Clone(s.onDestroyObservers)

		s.eventLoop.Post(func() {
			// the event loop is panic protected, thus separate the observer execution
			for _, f := range tmp {
				f()
			}
		})
		s.eventLoop.Post(func() {
			s.destroy()
		})

		s.eventLoop.Destroy()

		return true
	})

}

func (s *Scope) AddOnDestroyObserver(f func()) {
	s.destroyed.With(func(b bool) bool {
		s.onDestroyObservers = append(s.onDestroyObservers, f)
		return b
	})
}

// only for event loop
func (s *Scope) destroy() {
	s.cancelCtx()
	for _, component := range s.allocatedComponents {
		invokeDestructors(component)
	}

	s.channel.Store(NewPrintChannel()) // detach
	clear(s.allocatedComponents)
	//	clear(s.factories) // clearing this map would cause a data race, even though we use the factory as read-only
}

// only for event loop
func (s *Scope) handleSessionAssigned(evt ora.SessionAssigned) {
	s.sessionID = SessionID(evt.SessionID)
}

// only for event loop
func invokeDestructors(component allocatedComponent) {
	component.Window.viewRoot.Destroy()

	if closer, ok := component.Component.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			slog.Error("error on closing Component", slog.Any("err", err), slog.String("type", fmt.Sprintf("%T", component)))
		}
	}

	if destroyer, ok := component.Component.(Destroyable); ok {
		destroyer.Destroy()
	}
}
