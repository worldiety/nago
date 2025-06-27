// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type TDivider struct {
	frame   proto.Frame
	border  proto.Border
	padding proto.Padding
}

// HLineWithColor configures the TDivider to be used as a horizontal hairline divider, e.g. within a TVStack.
func HLineWithColor(c Color) TDivider {
	return TDivider{}.
		Border(Border{TopWidth: "1px", TopColor: c}).
		Frame(Frame{}.FullWidth()).
		Padding(Padding{}.Vertical(L16))

}

// HLine configures the TDivider to be used as a horizontal hairline divider, e.g. within a TVStack.
// The color is derived from the main color.
func HLine() TDivider {
	return TDivider{}.
		Border(Border{TopWidth: "1px", TopColor: M5}).
		Frame(Frame{}.FullWidth()).
		Padding(Padding{}.Vertical(L16))

}

// VLineWithColor configures a TDivider to be used as a vertical hairline divider, e.g. within a THStack.
func VLineWithColor(c Color) TDivider {
	return TDivider{}.
		Border(Border{LeftWidth: "1px", LeftColor: c}).
		Frame(Frame{}.FullHeight()).
		Padding(Padding{}.Horizontal(L16))

}

// VLine configures a TDivider to be used as a vertical hairline divider, e.g. within a THStack.
// The color is derived from main.
func VLine() TDivider {
	return TDivider{}.
		Border(Border{LeftWidth: "1px", LeftColor: M5}).
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

func (c TDivider) Render(ctx core.RenderContext) core.RenderNode {

	return &proto.Divider{
		Frame:   c.frame,
		Border:  c.border,
		Padding: c.padding,
	}
}
