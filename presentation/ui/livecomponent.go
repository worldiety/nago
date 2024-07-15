package ui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"reflect"
	"strconv"
)

type String = *Shared[string]
type EmbeddedSVG = *Shared[SVGSrc]
type EmbeddedAlignment = *Shared[Alignment]
type EmbeddedOrientation = *Shared[Orientation]
type EmbeddedElementSize = *Shared[ElementSize]
type EmbeddedContentAlignment = *Shared[ContentAlignment]
type EmbeddedItemsAlignment = *Shared[ItemsAlignment]
type EmbeddedNavigationComponent = *Shared[ora.NavigationComponent]
type Bool = *Shared[bool]
type Int = *Shared[int64]
type Float = *Shared[float64]

type SVGSrc = ora.SVG
type Alignment = ora.Alignment
type ElementSize = ora.ElementSize
type Orientation = ora.Orientation
type ContentAlignment = ora.ContentAlignment
type ItemsAlignment = ora.ItemsAlignment

// Allows sizes are sm, base, lg, xl and 2xl
type Size string

// Shared represents a shared state between client and server. Both sides share the same state connected through
// a message bus. We could use a comparable here, however components would not work due to function pointers.
type Shared[T any] struct {
	id    ora.Ptr
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

func (s *Shared[T]) render() ora.Property[T] {
	return ora.Property[T]{
		Ptr:   s.id,
		Value: s.v,
	}
}

func (s *Shared[T]) Iter(f func(T) bool) {
	f(s.v)
}

func (s *Shared[T]) AnyIter(f func(any) bool) {
	if c, ok := any(s.v).(core.View); ok {
		if isNil(c) {
			return
		}
	}
	f(s.v)
}

func (s *Shared[T]) ID() ora.Ptr {
	return s.id
}

func (s *Shared[T]) Unwrap() any {
	return s.v
}

func (s *Shared[T]) setValue(value string) error {
	return s.Parse(value)
}

func (s *Shared[T]) Parse(value string) error {
	s.SetDirty(true)

	switch any(s.v).(type) {
	case string:
		s.v = any(value).(T)
	case SVGSrc:
		s.v = any(value).(T)
	case ora.NamedColor:
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

func nextPtr() ora.Ptr {
	return core.NextPtr()
}

func nextToken() string {
	var tmp [16]byte
	_, err := rand.Read(tmp[:])
	if err != nil {
		panic(fmt.Errorf("unexpected crypto error: %w", err))
	}
	return hex.EncodeToString(tmp[:])
}

func renderFunc(lf *core.Func) ora.Property[ora.Ptr] {
	if lf.Nil() {
		return ora.Property[ora.Ptr]{}
	}

	return ora.Property[ora.Ptr]{
		Ptr:   lf.ID(), // TODO why is this not a Property or a Shared[func()]? it is a logical slot (itself a pointer) with a value set (again pointer)
		Value: lf.ID(),
	}
}

func isNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}
