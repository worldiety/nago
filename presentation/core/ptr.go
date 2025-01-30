package core

import (
	"go.wdy.de/nago/presentation/proto"
	"sync/atomic"
)

var nextFakePtr int64

func NextPtr() proto.Ptr {
	return proto.Ptr(atomic.AddInt64(&nextFakePtr, 1))
}
