package ui

import (
	"fmt"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
)

// TODO this is the wrong signature
type Iter[T any] func(yield func(T))

type SharedList[T any] struct {
	id     CID
	name   string
	values []T
	dirty  bool
	iter   Iter[T]
}

func NewSharedList[T any](name string) *SharedList[T] {
	return &SharedList[T]{
		id:   nextPtr(),
		name: name,
	}
}

func (s *SharedList[T]) Name() string {
	return s.name
}

// From enables an "always" dirty re-evaluation of the shared entries.
// This disables defacto all possible optimizations, however it is the most comfortable.
func (s *SharedList[T]) From(iter Iter[T]) {
	s.iter = iter
}

// deprecated
func (s *SharedList[T]) Unwrap() any {
	var zero T
	_, isLiveComponent := any(zero).(core.Component)

	if s.iter != nil {
		if isLiveComponent {
			var tmp []LiveComponent
			s.iter(func(t T) {
				tmp = append(tmp, any(t).(LiveComponent))
			})
			return tmp

		} else {
			var tmp []T
			s.iter(func(t T) {
				tmp = append(tmp, t)
			})
			return tmp
		}
	}

	if isLiveComponent {
		var tmp []LiveComponent
		for _, value := range s.values {
			tmp = append(tmp, any(value).(LiveComponent))
		}
		return slice.Of(tmp...)
	}
	return slice.Of(s.values...)
}

func (s *SharedList[T]) ID() CID {
	return s.id
}

func (s *SharedList[T]) Parse(v string) error {
	return fmt.Errorf("cannot set shared values by list!?: %v", v)
}

func (s *SharedList[T]) Dirty() bool {
	if s == nil {
		return false
	}
	if s.iter != nil {
		return true
	}

	return s.dirty
}

func (s *SharedList[T]) SetDirty(b bool) {
	if b && s == nil {
		panic("cannot set non-false to nil shared list")
	}
	if !b && s == nil {
		return
	}

	s.dirty = b
}

// deprecated wrong iter signature, use Iter
func (s *SharedList[T]) Each(f func(T)) {
	if s == nil {
		return
	}

	if s.iter != nil {
		s.iter(f)
	}

	for _, value := range s.values {
		f(value)
	}
}

func (s *SharedList[T]) AnyIter(f func(any) bool) {
	if s == nil {
		return
	}

	if s.iter != nil {
		s.iter(func(t T) {
			// return f(t)
			f(t) // todo update to proper iterator type
		})
	}

	for _, value := range s.values {
		if !f(value) {
			return
		}
	}
}

func (s *SharedList[T]) Iter(f func(T) bool) {
	if s == nil {
		return
	}

	if s.iter != nil {
		s.iter(func(t T) {
			// return f(t)
			f(t) // todo update to proper iterator type
		})
	}

	for _, value := range s.values {
		if !f(value) {
			return
		}
	}
}

// Has no effect if Source has been set.
func (s *SharedList[T]) Len() int {
	if s == nil {
		return 0
	}

	return len(s.values)
}

// Has no effect if Source has been set.
func (s *SharedList[T]) Append(t ...T) {
	if s.iter != nil {
		panic("cannot append data if Source has been set")
	}

	s.values = append(s.values, t...)
	s.dirty = true
}

// AppendFrom is like append but uses the given iter once. See also From for an always dirty yielding.
func (s *SharedList[T]) AppendFrom(iter Iter[T]) {
	iter(func(t T) {
		s.values = append(s.values, t)
	})

	s.dirty = true
}

// Remove removes the first comparable and matching entry. Has no effect if Source has been set.
func (s *SharedList[T]) Remove(t ...T) {
	if s.iter != nil {
		panic("cannot remove data if Source has been set")
	}

	anyRemoved := false
	tmp := make([]T, 0, len(s.values))
	for _, value := range s.values {
		doRemove := false
		for _, toRemove := range t {
			if any(value) == any(toRemove) { //hum
				doRemove = true
				anyRemoved = true
				break
			}
		}
		if !doRemove {
			tmp = append(tmp, value)
		}
	}

	s.values = tmp
	s.dirty = anyRemoved
}

// Clear removes any contained pointers and sets the length to 0. Has no effect if Source has been set.
func (s *SharedList[T]) Clear() {
	if s.iter != nil {
		panic("cannot clear data if Source has been set")
	}

	var zero T
	for i := range s.values {
		s.values[i] = zero
	}
	s.values = s.values[:0]
	s.dirty = true
}

func renderSharedListButtons(s *SharedList[*Button]) protocol.Property[[]protocol.Button] {
	res := protocol.Property[[]protocol.Button]{
		Ptr: s.id,
	}

	for _, value := range s.values {
		res.Value = append(res.Value, value.renderButton())
	}

	return res
}

func renderSharedListComponents(s *SharedList[core.Component]) protocol.Property[[]protocol.Component] {
	res := protocol.Property[[]protocol.Component]{
		Ptr: s.id,
	}

	for _, value := range s.values {
		res.Value = append(res.Value, value.Render())
	}

	return res
}

func renderSharedComponent(s *Shared[core.Component]) protocol.Property[protocol.Component] {
	res := protocol.Property[protocol.Component]{
		Ptr: s.id,
	}

	if s.v != nil {
		res.Value = s.v.Render()
	}

	return res
}
