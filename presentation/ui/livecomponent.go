package ui

import (
	"fmt"
	"go.wdy.de/nago/container/slice"
	"strconv"
	"sync/atomic"
)

type LiveComponent interface {
	ID() CID
	Type() string
	Properties() slice.Slice[Property]
}

type Property interface {
	// Name returns the actual protocol name of this property.
	Name() string
	// Dirty returns true, if the property has been changed.
	Dirty() bool
	value() any
	// ID returns the internal unique instance ID of this property which is used to identify it across process
	// boundaries.
	ID() CID
	setValue(v string) error // don't touch
	// SetDirty explicitly marks or unmarks this property as dirty.
	// This is done automatically, when updating the value.
	SetDirty(b bool)
}

type String = *Shared[string]
type EmbeddedSVG = *Shared[SVGSrc]
type Bool = *Shared[bool]
type Int = *Shared[int64]
type Float = *Shared[float64]

type SVGSrc string

// Allows sizes are sm, base, lg, xl and 2xl
type Size string

type SharedList[T any] struct {
	id     CID
	name   string
	values []T
	dirty  bool
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

func (s *SharedList[T]) value() any {
	var zero T
	if _, ok := any(zero).(LiveComponent); ok {
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

func (s *SharedList[T]) setValue(v string) error {
	return fmt.Errorf("cannot set shared values by list!?: %v", v)
}

func (s *SharedList[T]) Dirty() bool {
	if s == nil {
		return false
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

func (s *SharedList[T]) Each(f func(T)) {
	if s == nil {
		return
	}

	for _, value := range s.values {
		f(value)
	}
}

func (s *SharedList[T]) Len() int {
	if s == nil {
		return 0
	}

	return len(s.values)
}

func (s *SharedList[T]) Append(t ...T) {
	s.values = append(s.values, t...)
	s.dirty = true
}

func (s *SharedList[T]) Remove(t ...T) {
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

// Clear removes any contained pointers and sets the length to 0.
func (s *SharedList[T]) Clear() {
	var zero T
	for i := range s.values {
		s.values[i] = zero
	}
	s.values = s.values[:0]
	s.dirty = true
}

// Shared represents a shared state between client and server. Both sides share the same state connected through
// a message bus. We could use a comparable here, however components would not work due to function pointers.
type Shared[T any] struct {
	id    CID
	v     T
	dirty bool
	name  string
}

func NewShared[T any](name string) *Shared[T] {
	return &Shared[T]{
		id:   nextPtr(),
		name: name,
	}
}

func (s *Shared[T]) ID() CID {
	return s.id
}

func (s *Shared[T]) value() any {
	return s.v
}

func (s *Shared[T]) setValue(value string) error {
	switch any(s.v).(type) {
	case string:
		s.v = any(value).(T)
	case SVGSrc:
		s.v = any(value).(T)
	case Color:
		s.v = any(value).(T)
	case bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		s.v = any(b).(T)
	case float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		s.v = any(f).(T)
	case int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		s.v = any(i).(T)
	default:
		panic(fmt.Errorf("unsupported primitive property: %T", s.v))
	}

	return nil
}

func (s *Shared[T]) Name() string {
	return s.name
}

func (s *Shared[T]) Set(value T) {
	s.dirty = true
	s.v = value
}

func (s *Shared[T]) Get() T {
	return s.v
}

func (s *Shared[T]) Dirty() bool {
	return s.dirty
}

func (s *Shared[T]) SetDirty(b bool) {
	s.dirty = b
}

type CID int64

func (c CID) Nil() bool {
	return c == 0
}

var nextFakePtr int64

func nextPtr() CID {
	return CID(atomic.AddInt64(&nextFakePtr, 1))
}

func With[T any](t T, f func(t T)) T {
	f(t)
	return t
}

func IsDirty(dst LiveComponent) bool {
	if dst == nil {
		return false
	}

	dirty := false
	dst.Properties().Each(func(idx int, v Property) {
		dirty = dirty || v.Dirty()
	})

	if dirty {
		return true
	}

	Functions(dst, func(f *Func) {
		dirty = dirty || f.Dirty()
	})

	if dirty {
		return true
	}

	Children(dst, func(c LiveComponent) {
		dirty = dirty || IsDirty(c)
	})

	return dirty
}

func SetDirty(dst LiveComponent, dirty bool) {
	if dst == nil {
		return
	}

	dst.Properties().Each(func(idx int, v Property) {
		v.SetDirty(dirty)
	})

	Functions(dst, func(f *Func) {
		f.SetDirty(dirty)
	})

	Children(dst, func(c LiveComponent) {
		SetDirty(c, dirty)
	})

}

func Functions(c LiveComponent, f func(f *Func)) {
	c.Properties().Each(func(idx int, v Property) {
		if fun, ok := v.(*Func); ok {
			f(fun)
		}
	})
}

func Children(c LiveComponent, f func(c LiveComponent)) {
	c.Properties().Each(func(idx int, v Property) {
		switch t := v.(type) {
		case *Shared[LiveComponent]:
			f(t.Get())
		case *SharedList[LiveComponent]:
			t.Each(func(component LiveComponent) {
				f(component)
			})
		case Property:
			switch v := t.value().(type) {
			case slice.Slice[LiveComponent]:
				v.Each(func(idx int, v LiveComponent) {
					f(v)
				})
			case LiveComponent:
				f(v)
			}
		}
	})
}
