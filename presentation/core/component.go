package core

import (
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/presentation/protocol"
)

type Component interface {
	ID() protocol.Ptr
	Type() string
	Properties() slice.Slice[Property]
}
