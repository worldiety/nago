package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TWindowTitle struct {
	title string
}

func WindowTitle(title string) TWindowTitle {
	return TWindowTitle{title: title}
}

func (c TWindowTitle) Render(ctx core.RenderContext) ora.Component {
	return ora.WindowTitle{
		Type:  ora.WindowTitleT,
		Value: c.title,
	}
}
