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

type TFilledButton struct {
	fillColor Color
	textColor Color
	title     string
	preIcon   proto.SVG
	postIcon  proto.SVG
	frame     Frame
	action    func()
}

func FilledButton(fillColor Color, action func()) TFilledButton {
	return TFilledButton{fillColor: fillColor, action: action}
}

func (c TFilledButton) Title(text string) TFilledButton {
	c.title = text
	return c
}

func (c TFilledButton) PreIcon(svg core.SVG) TFilledButton {
	c.preIcon = proto.SVG(svg)
	return c
}

func (c TFilledButton) PostIcon(svg core.SVG) TFilledButton {
	c.postIcon = proto.SVG(svg)
	return c
}

func (c TFilledButton) TextColor(color Color) TFilledButton {
	c.textColor = color
	return c
}

func (c TFilledButton) Frame(frame Frame) TFilledButton {
	c.frame = frame
	return c
}

func (c TFilledButton) Render(context core.RenderContext) proto.Component {
	decView := customButton(c.fillColor, HStack(
		If(len(c.preIcon) != 0, Image().Embed(c.preIcon).Frame(Frame{}.Size(L16, L16))),
		If(c.title != "", btnTitle(c.title, c.textColor)),
		If(len(c.postIcon) != 0, Image().Embed(c.postIcon).Frame(Frame{}.Size(L16, L16))),
	).Action(c.action))

	if (c.title == "" && len(c.preIcon) != 0) || (c.title == "" && len(c.postIcon) != 0) {
		decView = decView.Frame(Frame{Width: L40, Height: L40}).Padding(Padding{}.Horizontal("0dp"))
	}

	var zero Frame
	if c.frame != zero {
		decView = decView.Frame(c.frame)
	}

	return decView.Render(context)
}

func btnTitle(text string, color Color) TText {
	t := Text(text).Font(Font{Size: L14, Weight: 500})
	if color != "" {
		t = t.Color(color)
	}

	return t
}

func customButton(fillColor Color, hstack THStack) DecoredView {
	return hstack.
		HoveredBackgroundColor(fillColor.WithTransparency(25)).
		PressedBackgroundColor(fillColor.WithTransparency(35)).
		PressedBorder(Border{}.Circle().Color("#00000000").Width(L2)).
		FocusedBorder(Border{}.Circle().Color("#ffffff").Width(L2)).
		Gap(L4).
		BackgroundColor(fillColor).
		Frame(Frame{Height: "2.375rem"}).
		Padding(Padding{}.Horizontal("1.125rem")).
		// add invisible default border, to avoid dimension changes,
		// note, that we need to fix that with frame and padding above
		Border(Border{}.Circle().Color("#00000000").Width(L2))
}
