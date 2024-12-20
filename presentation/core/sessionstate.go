package core

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

type TransientProperty interface {
	isTransient()
	dirty() bool
	setGeneration(generation int64)
}

type TransientState[T any] struct {
	id                    string
	value                 T
	valid                 bool
	mutex                 sync.Mutex
	generation            int64
	lastChangedGeneration int64
}

func (s *TransientState[T]) Get() T {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.value
}

func (s *TransientState[T]) Set(value T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	atomic.StoreInt64(&s.lastChangedGeneration, s.getGeneration())

	s.valid = true
	s.value = value
}

func (s *TransientState[T]) getGeneration() int64 {
	return atomic.LoadInt64(&s.generation)
}

func (s *TransientState[T]) setGeneration(g int64) {
	atomic.StoreInt64(&s.generation, g)
}

func (s *TransientState[T]) Valid() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.valid
}

func (s *TransientState[T]) dirty() bool {
	return atomic.LoadInt64(&s.lastChangedGeneration) >= s.getGeneration()
}

/*func (s *TransientState[T]) Mutex() *sync.Mutex {
	return &s.mutex
}*/

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
