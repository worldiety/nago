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
		Border(Border{TopWidth: "1px", TopColor: "A0"}).
		Frame(Frame{}.FullWidth()).
		Padding(Padding{}.Vertical(L16))

}

// VLine configures a TDivider to be used as a vertical hairline divider, e.g. within a THStack.
func VLine() TDivider {
	return TDivider{}.
		Border(Border{TopWidth: "1px", TopColor: "A0"}).
		Frame(Frame{}.FullHeight()).
		Padding(Padding{}.Horizontal(L16))

}

func (c TDivider) Padding(padding Padding) TDivider {
	c.padding = padding.ora()
	return c
}

func (c TDivider) Frame(frame Frame) TDivider {
	c.frame = frame.ora()
	return c
}

func (c TDivider) Border(border Border) TDivider {
	c.border = border.ora()
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
