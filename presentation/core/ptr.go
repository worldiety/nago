package core

import (
	"go.wdy.de/nago/presentation/ora"
	"sync/atomic"
)

var nextFakePtr int64

func NextPtr() ora.Ptr {
	return ora.Ptr(atomic.AddInt64(&nextFakePtr, 1))
}
