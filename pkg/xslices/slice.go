package xslices

import (
	"encoding/json"
	"iter"
	"slices"
)

// Slice is an immutable view into a conventional slice. Note, that only variables used for unmarshalling
// will implicitly use a pointer to change its own value.
type Slice[T any] struct {
	s []T
}

func New[T any](v ...T) Slice[T] {
	return Slice[T]{v}
}

func (s Slice[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.s)
}

func (s *Slice[T]) UnmarshalJSON(data []byte) error {
	s.s = nil // intentionally discard that reference to ensure that shared slices will not get accidentally corrupted
	return json.Unmarshal(data, &s.s)
}

func (s Slice[T]) Len() int {
	return len(s.s)
}

func (s Slice[T]) All() iter.Seq[T] {
	return slices.Values(s.s)
}

// Clone returns a shallow clone of the underlying slice.
func (s Slice[T]) Clone() []T {
	return slices.Clone(s.s)
}

// Append copies all values into a new underlying slice and returns a new immutable slice.
func (s Slice[T]) Append(v ...T) Slice[T] {
	alloc := make([]T, len(s.s)+len(v))
	copy(alloc, s.s)
	copy(alloc[len(s.s):], v)
	return Slice[T]{alloc}
}
