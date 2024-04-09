package slice

import "encoding/json"

// deprecated: iter.Seq is the same?
// Slice represents an immutable slice.
type Slice[T any] struct {
	slice []T
}

func (s *Slice[T]) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &s.slice)
}

func (s Slice[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.slice)
}

// Of takes the ownership of the given slice.
func Of[T any](v ...T) Slice[T] {
	return Slice[T]{slice: v}
}

// Append reuses the underlying slice and returns a new slice.
func (s Slice[T]) Append(v T) Slice[T] {
	return Slice[T]{
		slice: append(s.slice, v),
	}
}

// AppendAll reuses the underlying slice and returns a new slice.
func (s Slice[T]) AppendAll(v ...T) Slice[T] {
	return Slice[T]{
		slice: append(s.slice, v...),
	}
}

// Each is the iterator closure.
func (s Slice[T]) Each(f func(idx int, v T)) {
	for i, t := range s.slice {
		f(i, t)
	}
}

// At returns the value at the given index or panics.
func (s Slice[T]) At(idx int) T {
	return s.slice[idx]
}

// Len returns the length of the slice
func (s Slice[T]) Len() int {
	return len(s.slice)
}

// UnsafeUnwrap is unsafe and returns the underlying slice without any coping.
// Usually, you never should use this to optimize things.
// Never use it to modify the underlying slice.
func UnsafeUnwrap[T any](s Slice[T]) []T {
	return s.slice
}

// Map maps each entry of the slice to a new entry. This causes a single
// full copy of the slice. Thus, updating one element is equal to updating everything.
func Map[From, To any](s Slice[From], f func(idx int, v From) To) Slice[To] {
	clone := make([]To, s.Len())
	for i, from := range s.slice {
		clone[i] = f(i, from)
	}

	return Slice[To]{slice: clone}
}
