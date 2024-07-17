package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TSpacer struct {
	alignment       ora.Alignment
	backgroundColor ora.Color
	frame           ora.Frame
	border          ora.Border
}

func Spacer() TSpacer {
	return TSpacer{}
}

func (c TSpacer) Alignment() ora.Alignment {
	return c.alignment
}

func (c TSpacer) SetAlignment(alignment ora.Alignment) {
	c.alignment = alignment
}

func (c TSpacer) BackgroundColor() ora.Color {
	return c.backgroundColor
}

func (c TSpacer) SetBackgroundColor(backgroundColor ora.Color) {
	c.backgroundColor = backgroundColor
}

func (c TSpacer) Frame() ora.Frame {
	return c.frame
}

func (c TSpacer) SetFrame(frame ora.Frame) {
	c.frame = frame
}

func (c TSpacer) Border() ora.Border {
	return c.border
}

func (c TSpacer) SetBorder(border ora.Border) {
	c.border = border
}

func (c TSpacer) Render(ctx core.RenderContext) ora.Component {

	return ora.Spacer{
		Type:            ora.SpacerT,
		Frame:           c.frame,
		BackgroundColor: c.backgroundColor,
		Border:          c.border,
	}
}
