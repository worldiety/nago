package uilegacy

import "go.wdy.de/nago/presentation/core"

type Container interface {
	Children() *SharedList[core.View]
}
