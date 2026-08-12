package flowchart

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// Menu represents an optional menu in the flowchart.
type Menu struct {
	Position Point
	Content  core.View
}

func (m Menu) render(ctx core.RenderContext) proto.FlowChartMenu {
	var content proto.Component
	if m.Content != nil {
		content = m.Content.Render(ctx)
	}

	return proto.FlowChartMenu{
		Position: m.Position.Ora(),
		Content:  content,
	}
}
