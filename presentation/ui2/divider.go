package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TDivider struct {
	frame   ora.Frame
	border  ora.Border
	padding ora.Padding
}

// HLine configures the TDivider to be used as a horizontal hairline divider, e.g. within a TVStack.
func HLine() TDivider {
	return TDivider{}.
		Border(ora.Border{TopWidth: "1px", TopColor: "A0"}).
		Frame(ora.Frame{}.FullWidth()).
		Padding(ora.Padding{}.Vertical(ora.L16))

}

// VLine configures a TDivider to be used as a vertical hairline divider, e.g. within a THStack.
func VLine() TDivider {
	return TDivider{}.
		Border(ora.Border{TopWidth: "1px", TopColor: "A0"}).
		Frame(ora.Frame{}.FullHeight()).
		Padding(ora.Padding{}.Horizontal(ora.L16))

}

func (c TDivider) Padding(padding ora.Padding) TDivider {
	c.padding = padding
	return c
}

func (c TDivider) Frame(frame ora.Frame) TDivider {
	c.frame = frame
	return c
}

func (c TDivider) Border(border ora.Border) TDivider {
	c.border = border
	return c
}

func (c TDivider) Render(ctx core.RenderContext) ora.Component {

	return ora.Divider{
		Type:    ora.DividerT,
		Frame:   c.frame,
		Border:  c.border,
		Padding: c.padding,
	}
}
