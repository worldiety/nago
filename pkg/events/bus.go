// Package events provides a simple async in-process event distribution system.
// Keep the usage of this package to a minimum, because it may be an indicator of a poorly designed and
// maintainable architecture. It is easy to create invisible and fragile workflows based on complex event chains.
// Consider using a workflow engine.
// So, when to use this package at all? The following aspects may be relevant:
//   - a lot of build-in use cases generate (domain) specific events. Because these basic nago systems must be
//     usually declared and injected into actual domain code, there is a (runtime) dependency cycle. This cycle can
//     be made dynamically, however there are racy questions to that. A simple bus looks more elegant, than other
//     observer/listener patterns, thus it may be the lesser of two evils.
//   - nago itself has cyclic dependencies at its core like the mail system which needs the user subject and
//     the user self-service functions need the mail system.
package events

import (
	"fmt"
	"log/slog"
	"reflect"
	"runtime/debug"
	"sync"
)

type busSubscriberOptions struct {
	typeDef reflect.Type
}

type SubscriberOption interface {
	apply(*busSubscriberOptions)
}

type subOptFunc func(*busSubscriberOptions)

func (f subOptFunc) apply(opts *busSubscriberOptions) {
	f(opts)
}

// EventBus is a basic contract to publish messages and to subscribe for them.
type EventBus interface {
	// Publish issues a new event to all subscribers. Depending on the implementation, this may
	// sync or async. If evt is the nil interface, nothing is published.
	Publish(evt any)
	Subscribe(fn func(evt any), opts ...SubscriberOption) (close func())
}

func TypeFor[T any]() SubscriberOption {
	return subOptFunc(func(opts *busSubscriberOptions) {
		opts.typeDef = reflect.TypeFor[T]()
	})
}

// NewEventBus creates a new default async event bus. Each invocation to a subscriber spawns a new go routine, which is
// expensive but free of stalls and deadlocks. See package documentation [events] to better understand when
// to use this pattern.
func NewEventBus() EventBus {
	return &asyncEvents{
		typeObservers: make(map[reflect.Type]map[hnd]func(evt any)),
		anyObserver:   make(map[hnd]func(evt any)),
	}
}

type hnd int
type asyncEvents struct {
	mutex         sync.RWMutex
	typeObservers map[reflect.Type]map[hnd]func(evt any)
	anyObserver   map[hnd]func(evt any)
	lastHnd       hnd
}

func (a *asyncEvents) Publish(evt any) {
	if evt == nil {
		return
	}

	refType := reflect.TypeOf(evt)
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	for _, fn := range a.anyObserver {
		a.spawn(evt, fn)
	}

	subMap := a.typeObservers[refType]
	for _, fn := range subMap {
		a.spawn(evt, fn)
	}
}

func (a *asyncEvents) spawn(evt any, fn func(evt any)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(r)
				debug.PrintStack()
				slog.Error("panic in event bus")
			}
		}()

		fn(evt)
	}()
}

func (a *asyncEvents) Subscribe(fn func(evt any), opts ...SubscriberOption) (close func()) {
	var cfg busSubscriberOptions
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.lastHnd++
	myHnd := a.lastHnd
	if cfg.typeDef == nil {
		a.anyObserver[myHnd] = fn
		return func() {
			a.mutex.Lock()
			defer a.mutex.Unlock()

			delete(a.anyObserver, myHnd)
		}
	}

	subMap, ok := a.typeObservers[cfg.typeDef]
	if !ok {
		subMap = make(map[hnd]func(evt any))
		a.typeObservers[cfg.typeDef] = subMap
	}

	subMap[myHnd] = fn
	return func() {
		a.mutex.Lock()
		defer a.mutex.Unlock()
		delete(subMap, myHnd)
	}

}

// SubscribeFor just listens for events of the according type which is likely more efficient, than registering
// for any events.
func SubscribeFor[T any](evtBus EventBus, fn func(evt T)) (close func()) {
	return evtBus.Subscribe(func(evt any) {
		fn(evt.(T))
	}, TypeFor[T]())
}
