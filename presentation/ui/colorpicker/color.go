package colorpicker

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type TColor struct {
	color ui.Color
}

func Color(color ui.Color) TColor {
	return TColor{color: color}
}

func (c TColor) Render(ctx core.RenderContext) core.RenderNode {
	return renderColor(nil, c.color).Render(ctx)
}
