package core

import (
	"go.wdy.de/nago/presentation/ora"
)

type Property interface {
	// ptrId returns the internal unique instance ID of this property which is used to identify it across process
	// boundaries.
	ptrId() ora.Ptr

	parse(v any) error
	getGeneration() int64
	setGeneration(g int64)
	clearObservers()
	destroy()
	isDestroyed() bool
	dirty() bool
}
