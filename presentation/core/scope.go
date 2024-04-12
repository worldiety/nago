package core

import (
	"fmt"
	"go.wdy.de/nago/presentation/protocol"
	"io"
	"log/slog"
	"sync/atomic"
	"time"
)

type Destroyable interface {
	Destroy()
}

type allocatedComponent struct {
	Component   Component
	RenderState *RenderState
}

type ComponentFactory func(*Scope, protocol.NewComponentRequested) Component

// A Scope manage its own area of associated pointers. A Pointer must be only unique per Scope.
// Resolving or keeping pointers outside a scope is inherently unsafe (e.g. a lookup map).
// A Scope can hold an arbitrary amount of components, created by an arbitrary amount of factories.
// This is intentional. E.g. a mobile app can create multiple instances of the same (or different)
// "pages" and those pages may communicate immediately with each other (observer etc.) without
// causing race conditions.
// Each Scope has a single event loop to guarantee race free event processing.
type Scope struct {
	id                  protocol.ScopeID
	factories           map[protocol.ComponentFactoryId]ComponentFactory
	allocatedComponents map[protocol.Ptr]allocatedComponent
	lifetime            time.Duration
	endOfLifeAt         atomic.Pointer[time.Time]
	channel             AtomicRef[Channel]
	chanDestructor      AtomicRef[func()]
	eventLoop           *EventLoop
}

func NewScope(id protocol.ScopeID, lifetime time.Duration, factories map[protocol.ComponentFactoryId]ComponentFactory) *Scope {
	s := &Scope{
		id:                  id,
		factories:           factories,
		allocatedComponents: map[protocol.Ptr]allocatedComponent{},
		lifetime:            lifetime,
		eventLoop:           NewEventLoop(),
	}

	s.channel.Store(NopChannel{})
	s.Tick()

	return s
}

// Connect attaches the given channel to this Scope immediately. There must be exact 1 Scope per Channel.
// The use case is, that Scopes can be transferred from one channel to another easily.
func (s *Scope) Connect(c Channel) {
	if c == nil {
		c = NopChannel{}
	}

	slog.Info("scope connected to channel", slog.String("id", string(s.id)), slog.String("channel", fmt.Sprintf("%T: %#v", c, c)))

	s.chanDestructor.With(func(destructor func()) func() {
		s.channel.Store(c)

		if destructor != nil {
			destructor()
		}

		return c.Subscribe(func(msg []byte) error {
			return s.handleMessage(msg)
		})
	})
}

func (s *Scope) handleMessage(buf []byte) error {
	s.Tick()

	t, err := protocol.Unmarshal(buf)
	if err != nil {
		return err
	}

	s.eventLoop.Post(func() {
		s.handleEvent(t)
	})

	return nil
}

func (s *Scope) Publish(evt protocol.Event) {
	if err := s.channel.Load().Publish(protocol.Marshal(evt)); err != nil {
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
func (s *Scope) sendAck(id protocol.RequestId) {
	if id == 0 {
		return
	}

	s.Publish(protocol.Acknowledged{
		Type:      protocol.AcknowledgedT,
		RequestId: id,
	})
}

// only for event loop
func (s *Scope) render(requestId protocol.RequestId, component Component) protocol.ComponentInvalidated {
	// TODO
	return protocol.ComponentInvalidated{
		Type:      protocol.ComponentInvalidatedT,
		RequestId: requestId,
		Component: nil,
	}
}

// Destroy frees all allocated components and removes factory pointers.
// The scope is of no use afterward.
// Do never call this from the event loop.
func (s *Scope) Destroy() {
	s.eventLoop.Post(func() {
		s.destroy()
	})

	s.eventLoop.Shutdown()
}

// only for event loop
func (s *Scope) destroy() {
	for _, component := range s.allocatedComponents {
		invokeDestructors(component)
	}

	s.channel.Store(NewPrintChannel()) // detach
	clear(s.allocatedComponents)
	clear(s.factories)
}

// only for event loop
func invokeDestructors(component any) {
	if closer, ok := component.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			slog.Error("error on closing Component", slog.Any("err", err), slog.String("type", fmt.Sprintf("%T", component)))
		}
	}

	if destroyer, ok := component.(Destroyable); ok {
		destroyer.Destroy()
	}
}
