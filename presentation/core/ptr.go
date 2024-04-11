package core

import (
	"go.wdy.de/nago/presentation/protocol"
	"sync/atomic"
)

var nextFakePtr int64

func NextPtr() protocol.Ptr {
	return protocol.Ptr(atomic.AddInt64(&nextFakePtr, 1))
}
