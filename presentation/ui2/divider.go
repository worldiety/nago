package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Divider struct {
	frame   ora.Frame
	border  ora.Border
	padding ora.Padding
	with    func(divider *Divider)
}

func NewDivider(with func(*Divider)) *Divider {
	c := &Divider{
		with: with,
	}

	return c
}

// HDivider configures the Divider to be used as a horizontal divider, e.g. within a TVStack.
func HDivider() *Divider {
	return NewDivider(func(divider *Divider) {
		var border ora.Border
		border.TopWidth = "1px"
		border.TopColor = "#00000019"
		divider.SetBorder(border)
		divider.SetFrame(ora.Frame{}.FullWidth())
		divider.SetPadding(ora.Padding{}.Vertical("0.25rem"))
	})
}

// VDivider configures a Divider to be used as a vertical divider, e.g. within a THStack.
func VDivider() *Divider {
	return NewDivider(func(divider *Divider) {
		var border ora.Border
		border.LeftWidth = "1px"
		border.LeftColor = "#00000019"
		divider.SetBorder(border)
		divider.SetFrame(ora.Frame{}.FullHeight())
		divider.SetPadding(ora.Padding{}.Horizontal("0.25rem"))
	})
}

func (c *Divider) Padding() ora.Padding {
	return c.padding
}

func (c *Divider) SetPadding(padding ora.Padding) {
	c.padding = padding
}

func (c *Divider) Frame() ora.Frame {
	return c.frame
}

func (c *Divider) SetFrame(frame ora.Frame) {
	c.frame = frame
}

func (c *Divider) Border() ora.Border {
	return c.border
}

func (c *Divider) SetBorder(border ora.Border) {
	c.border = border
}

func (c *Divider) Render(ctx core.RenderContext) ora.Component {
	if c.with != nil {
		c.with(c)
	}

	return ora.Divider{
		Type:    ora.DividerT,
		Frame:   c.frame,
		Border:  c.border,
		Padding: c.padding,
	}
}
