package core

import (
	"fmt"
	"log/slog"
	"sync"
)

type TransientProperty interface {
	isTransient()
}

type TransientState[T any] struct {
	id    string
	value T
	valid bool
	mutex sync.Mutex
}

func (s *TransientState[T]) Get() T {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.value
}

func (s *TransientState[T]) Set(value T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.valid = true
	s.value = value
}

func (s *TransientState[T]) Valid() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.valid
}

func (s *TransientState[T]) Mutex() *sync.Mutex {
	return &s.mutex
}

func (s *TransientState[T]) isTransient() {}

// TransientStateOf allocates a State which is held on the Scope (or connection/session level). This
// will be kept until the Scope is destroyed. Thus, be very careful not to leak too long and too much.
// Only put stuff here, which is lightweight and must survive different root views.
func TransientStateOf[T any](wnd Window, id string) *TransientState[T] {
	w := wnd.(*scopeWindow)
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if id == "" {
		panic("empty id is not allowed, consider using AutoState instead")
	}
	some, ok := w.parent.statesById[id]
	if ok {
		if found, ok := some.(*TransientState[T]); ok {
			return found
		}
		var zero T
		slog.Error("restored transient state does not match expected state type", "expected", fmt.Sprintf("%T", zero), "got", fmt.Sprintf("%T", some))
	}

	state := &TransientState[T]{
		id:    id,
		valid: false,
	}

	w.parent.statesById[id] = state

	return state
}
