package pubsub

import (
	"go.wdy.de/nago/pkg/xmaps"
	"go.wdy.de/nago/pkg/xreflect"
	"log/slog"
	"runtime/debug"
	"sync/atomic"
)

type fnHnd = int64

type PubSub struct {
	observers *xmaps.ConcurrentMap[xreflect.TypeID, *xmaps.ConcurrentMap[fnHnd, func(any)]]
	lastHnd   fnHnd
}

func NewPubSub() *PubSub {
	return &PubSub{
		observers: xmaps.NewConcurrentMap[xreflect.TypeID, *xmaps.ConcurrentMap[fnHnd, func(any)]](),
	}
}

// Publish creates a typeid from the given message and sends it to any registered subscriber. Each subscriber
// will be invoked from a new goroutine to ensure, that no deadlocks or stalls can occur. This is not as efficient,
// as other solutions, but this behavior will be correct in any case. Nil messages are ignored. Do not use
// pointer types for messages, due to ownership questions and due to the danger of sending typed nil values around
// which may be unexpected by subscribers.
func (p *PubSub) Publish(m any) {
	id, ok := xreflect.TypeIDFrom(m)
	if !ok {
		return
	}

	subscribers, ok := p.observers.Load(id)
	if !ok {
		return
	}

	for _, subscriber := range subscribers.All {
		go func() {
			if r := recover(); r != nil {
				slog.Error("PubSub panic on publish message", "err", r, "msg", m, "subscriber", subscriber)
				debug.PrintStack()
			}

			subscriber(m)
		}()
	}
}

// Subscribe performs a subscription for the given TypeID. If any of such message is publish, the subscriber will
// be invoked from a distinct goroutine. If the subscriber panics, the spawned goroutine is recovered and the
// incident is printed.
// The process is kept alive and any other subscribers will receive the event as usual.
func (p *PubSub) Subscribe(id xreflect.TypeID, fn func(value any)) (unsubscribe func()) {
	hnd := atomic.AddInt64(&p.lastHnd, 1)

	typeObservers, ok := p.observers.Load(id)
	if !ok {
		// perform a kind of double-check-idiom to avoid usually redundant allocations for 1+ subscriptions
		typeObservers, _ = p.observers.LoadOrStore(id, xmaps.NewConcurrentMap[fnHnd, func(any)]())
	}

	typeObservers.Store(hnd, fn)

	return func() {
		typeObservers.Delete(hnd)
	}
}
