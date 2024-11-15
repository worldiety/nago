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

func (b *Builder[T]) Last() (T, bool) {
	var zero T
	if len(b.tmp) == 0 {
		return zero, false
	}

	return b.tmp[len(b.tmp)-1], true
}

func (b *Builder[T]) First() (T, bool) {
	var zero T
	if len(b.tmp) == 0 {
		return zero, false
	}

	return b.tmp[0], true
}

func (b *Builder[T]) RemoveLast() (T, bool) {
	var zero T
	if len(b.tmp) == 0 {
		return zero, false
	}

	v := b.tmp[len(b.tmp)-1]
	b.tmp[len(b.tmp)-1] = zero
	b.tmp = b.tmp[:len(b.tmp)-1]

	return v, true
}
