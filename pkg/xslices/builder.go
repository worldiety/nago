package xslices

import (
	"iter"
	"slices"
)

type Builder[T any] struct {
	tmp []T
}

func NewBuilder[T any]() *Builder[T] {
	return &Builder[T]{}
}

func (b *Builder[T]) Append(v ...T) *Builder[T] {
	b.tmp = append(b.tmp, v...)
	return b
}

func (b *Builder[T]) Len() int {
	return len(b.tmp)
}

// Collect returns a mutable copied slice of the current state. The returned slice can be modified freely and
// future appends will not corrupt any prior returned slices.
func (b *Builder[T]) Collect() []T {
	return slices.Clone(b.tmp)
}

// All returns a read only view to iterate on of the current state.
func (b *Builder[T]) All() iter.Seq[T] {
	return slices.Values(b.tmp)
}
