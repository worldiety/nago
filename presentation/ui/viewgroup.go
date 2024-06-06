package ui

import "go.wdy.de/nago/presentation/core"

type Container interface {
	Children() *SharedList[core.Component]
}
