package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TSpacer struct {
	backgroundColor ora.Color
	frame           ora.Frame
	border          ora.Border
}

// Spacer is used in VStack or HStack to grow and shrink as required.
func Spacer() TSpacer {
	return TSpacer{}
}

func (c TSpacer) BackgroundColor(backgroundColor Color) TSpacer {
	c.backgroundColor = backgroundColor.ora()
	return c
}

func (c TSpacer) Frame(frame Frame) TSpacer {
	c.frame = frame.ora()
	return c
}

func (c TSpacer) Border(border Border) {
	c.border = border.ora()
}

func (c TSpacer) Render(ctx core.RenderContext) ora.Component {

	return ora.Spacer{
		Type:            ora.SpacerT,
		Frame:           c.frame,
		BackgroundColor: c.backgroundColor,
		Border:          c.border,
	}
}
