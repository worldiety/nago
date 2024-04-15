package ui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
	"strconv"
)

// deprecated
type LiveComponent = core.Component

// deprecated
type Property = core.Property

type String = *Shared[string]
type EmbeddedSVG = *Shared[SVGSrc]
type Bool = *Shared[bool]
type Int = *Shared[int64]
type Float = *Shared[float64]

type SVGSrc = protocol.SVG

// Allows sizes are sm, base, lg, xl and 2xl
type Size string

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

func (s *Shared[T]) Render() protocol.Property[T] {
	return protocol.Property[T]{
		Ptr:   s.id,
		Value: s.v,
	}
}

func (s *Shared[T]) Iter(f func(T) bool) {
	f(s.v)
}

func (s *Shared[T]) ID() CID {
	return s.id
}

func (s *Shared[T]) Unwrap() any {
	return s.v
}

func (s *Shared[T]) setValue(value string) error {
	return s.Parse(value)
}

func (s *Shared[T]) Parse(value string) error {
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

type CID = protocol.Ptr

func nextPtr() CID {
	return core.NextPtr()
}

func With[T any](t T, f func(t T)) T {
	f(t)
	return t
}

func nextToken() string {
	var tmp [16]byte
	_, err := rand.Read(tmp[:])
	if err != nil {
		panic(fmt.Errorf("unexpected crypto error: %w", err))
	}
	return hex.EncodeToString(tmp[:])
}
