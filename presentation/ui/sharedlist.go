package ui

import (
	"fmt"
	"go.wdy.de/nago/pkg/iter"
	slices2 "go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type SharedList[T any] struct {
	id           ora.Ptr
	name         string
	values       []T
	dirty        bool
	iter         iter.Seq[T]
	frozen       bool
	frozenValues []T
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
func (s *SharedList[T]) From(it iter.Seq[T]) {
	s.iter = it
}

func (s *SharedList[T]) render() ora.Property[[]T] {
	return ora.Property[[]T]{
		Ptr:   s.id,
		Value: slices2.Collect(s.Iter),
	}
}

func (s *SharedList[T]) ID() ora.Ptr {
	return s.id
}

func (s *SharedList[T]) Parse(v string) error {
	return fmt.Errorf("cannot set shared values by list!?: %v", v)
}

func (s *SharedList[T]) Dirty() bool {
	if s == nil {
		return false
	}
	if s.frozen {
		return s.dirty
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

	if s.frozen {
		s.dirty = b
		return
	}

	if !b && s == nil {
		return
	}

	s.dirty = b
}

func (s *SharedList[T]) AnyIter(f func(any) bool) {
	if s == nil {
		return
	}

	if s.frozen {
		for _, value := range s.frozenValues {
			if !f(value) {
				return
			}
		}

		return
	}

	if s.iter != nil {
		s.iter(func(t T) bool {
			return f(t)
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

	if s.frozen {
		for _, value := range s.frozenValues {
			if !f(value) {
				return
			}
		}

		return
	}

	if s.iter != nil {
		s.iter(func(t T) bool {
			return f(t)
		})
	}

	for _, value := range s.values {
		if !f(value) {
			return
		}
	}
}

func (s *SharedList[T]) Freeze() {
	if s.frozen {
		panic(fmt.Errorf("already frozen shared list %v", s.id))
	}

	tmp := make([]T, 0, len(s.values))

	//fmt.Printf("freezing shared list: %v\n", s.id)
	s.Iter(func(t T) bool {
		tmp = append(tmp, t)
		return true
	})

	s.frozen = true
	s.dirty = true
	s.frozenValues = tmp
}

func (s *SharedList[T]) Unfreeze() {
	if !s.frozen {
		panic(fmt.Errorf("already unfrozen shared list %v", s.id))
	}

	//fmt.Printf("unfreezing shared list: %v\n", s.id)

	s.frozen = false
	s.frozenValues = nil
}

// Has no effect if Source has been set.
func (s *SharedList[T]) Len() int {
	if s == nil {
		return 0
	}

	if s.frozen {
		return len(s.frozenValues)
	}

	return len(s.values)
}

// Has no effect if Source has been set.
func (s *SharedList[T]) Append(t ...T) {
	if s.iter != nil {
		panic("cannot append data if Source has been set")
	}

	if s.frozen {
		panic("cannot append data if frozen")
	}

	s.values = append(s.values, t...)
	s.dirty = true
}

// AppendFrom is like append but uses the given iter once. See also From for an always dirty yielding.
func (s *SharedList[T]) AppendFrom(it iter.Seq[T]) {
	if s.frozen {
		panic("cannot append data if frozen")
	}

	it(func(t T) bool {
		s.values = append(s.values, t)
		return true
	})

	s.dirty = true
}

// Remove removes the first comparable and matching entry. Has no effect if Source has been set.
func (s *SharedList[T]) Remove(t ...T) {
	if s.iter != nil {
		panic("cannot remove data if Source has been set")
	}

	if s.frozen {
		panic("cannot remove data if frozen")
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

// Clear removes any contained pointers and sets the length to 0. If a source iter has been set, the iter is removed.
func (s *SharedList[T]) Clear() {
	if s.frozen {
		panic("cannot clear data if frozen")
	}

	if s.iter != nil {
		s.iter = nil
		s.dirty = true
		return
	}

	var zero T
	for i := range s.values {
		s.values[i] = zero
	}
	s.values = s.values[:0]
	s.dirty = true
}

func renderSharedListButtons(s *SharedList[*Button]) ora.Property[[]ora.Button] {
	res := ora.Property[[]ora.Button]{
		Ptr: s.id,
	}

	s.Iter(func(button *Button) bool {
		res.Value = append(res.Value, button.renderButton())
		return true
	})

	return res
}

func renderSharedListMenuEntries(s *SharedList[*MenuEntry]) ora.Property[[]ora.MenuEntry] {
	res := ora.Property[[]ora.MenuEntry]{
		Ptr: s.id,
	}

	s.Iter(func(entry *MenuEntry) bool {
		res.Value = append(res.Value, entry.renderMenuEntry())
		return true
	})

	return res
}

func renderSharedListComponents(s *SharedList[core.Component]) ora.Property[[]ora.Component] {
	res := ora.Property[[]ora.Component]{
		Ptr: s.id,
	}

	s.Iter(func(value core.Component) bool {
		res.Value = append(res.Value, value.Render())
		return true
	})

	return res
}

func renderSharedListComponentsFlat(s *SharedList[core.Component]) []ora.Component {
	var res []ora.Component

	s.Iter(func(value core.Component) bool {
		res = append(res, value.Render())
		return true
	})

	return res
}

func renderSharedComponent(s *Shared[core.Component]) ora.Property[ora.Component] {
	res := ora.Property[ora.Component]{
		Ptr: s.id,
	}

	if s.v != nil {
		res.Value = s.v.Render()
	}

	return res
}

func renderSharedNavigationComponent(n *Shared[*NavigationComponent]) ora.Property[ora.NavigationComponent] {
	res := ora.Property[ora.NavigationComponent]{
		Ptr: n.id,
	}

	if n.v != nil {
		res.Value = n.v.renderNavigationComponent()
	}

	return res
}
