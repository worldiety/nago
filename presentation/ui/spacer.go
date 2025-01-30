package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// FixedSpacer returns an empty view with the given dimensions.
func FixedSpacer(width, height Length) core.View {
	return VStack(
		// double wrap, to trick the CSS flexbox (mis) behavior
		VStack().Frame(Frame{Width: width, Height: height}),
	)
}

type TSpacer struct {
	backgroundColor proto.Color
	frame           proto.Frame
	border          proto.Border
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

func (c TSpacer) Render(ctx core.RenderContext) core.RenderNode {

	return &proto.Spacer{
		Frame:           c.frame,
		BackgroundColor: c.backgroundColor,
		Border:          c.border,
	}
}
