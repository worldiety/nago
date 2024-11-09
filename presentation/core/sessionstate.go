package core

import (
	"fmt"
	"log/slog"
)

type TransientProperty interface {
	isTransient()
}

type TransientState[T any] struct {
	id    string
	value T
	valid bool
}

func (s *TransientState[T]) Get() T {
	return s.value
}

func (s *TransientState[T]) Set(value T) {
	s.value = value
}

func (s *TransientState[T]) Valid() bool {
	return s.valid
}

func (s *TransientState[T]) isTransient() {}

// TransientStateOf allocates a State which is held on the Scope (or connection/session level). This
// will be kept until the Scope is destroyed. Thus, be very careful not to leak too long and too much.
// Only put stuff here, which is lightweight and must survive different root views.
func TransientStateOf[T any](wnd Window, id string) *TransientState[T] {
	w := wnd.(*scopeWindow)
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
