package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
)

func NewView[Entity data.Aggregate[ID], ID data.IDType](wnd core.Window, opts *Options[Entity, ID]) core.View {
}
