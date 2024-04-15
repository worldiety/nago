package core

import "go.wdy.de/nago/presentation/protocol"

type Property interface {
	// Name returns the actual protocol name of this property.
	Name() string
	// Dirty returns true, if the property has been changed.
	Dirty() bool

	// deprecated: usage unclear, need a type assertion anyway why not assert the proper property type?
	Unwrap() any
	// ID returns the internal unique instance ID of this property which is used to identify it across process
	// boundaries.
	ID() protocol.Ptr

	Parse(v string) error
	// SetDirty explicitly marks or unmarks this property as dirty.
	// This is done automatically, when updating the value.
	SetDirty(b bool)
}

type Iterable[T any] interface {
	// Iter provides an iterator of iter.Seq[T]
	Iter(yield func(T) bool)
}
