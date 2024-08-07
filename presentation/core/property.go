package core

import (
	"go.wdy.de/nago/presentation/ora"
)

type Property interface {
	// ID returns the internal unique instance ID of this property which is used to identify it across process
	// boundaries.
	ID() ora.Ptr

	parse(v any) error
	getGeneration() int64
}
