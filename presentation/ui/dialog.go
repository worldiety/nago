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

type TDialog struct {
	uri              core.URI
	dlg              proto.VStack
	preBody          core.View
	body             core.View
	footer           core.View
	title            core.View
	titleX           core.View
	alignment        Alignment
	modalPadding     Padding
	frame            Frame
	disableBoxLayout bool
}

func Dialog(body core.View) TDialog {
	return TDialog{
		frame: Frame{Width: L400, MaxWidth: Full, MaxHeight: "calc(100dvh - 8rem)"},
		body:  body,
	}
}

func (c TDialog) Title(title core.View) TDialog {
	c.title = title
	return c
}

func (c TDialog) PreBody(v core.View) TDialog {
	c.preBody = v
	return c
}

func (c TDialog) TitleX(x core.View) TDialog {
	c.titleX = x
	return c
}

func (c TDialog) Footer(footer core.View) TDialog {
	c.footer = footer
	return c
}

func (c TDialog) Alignment(alignment Alignment) TDialog {
	c.alignment = alignment
	return c
}

func (c TDialog) ModalPadding(padding Padding) TDialog {
	c.modalPadding = padding
	return c
}

func (c TDialog) Frame(frame Frame) TDialog {
	c.frame = frame
	return c
}

func (c TDialog) WithFrame(fn func(Frame) Frame) TDialog {
	c.frame = fn(c.frame)
	return c
}

func (c TDialog) DisableBoxLayout(b bool) TDialog {
	c.disableBoxLayout = b
	return c
}

func (c TDialog) Render(ctx core.RenderContext) proto.Component {
	colors := core.Colors[Colors](ctx.Window())

	stack := VStack(
		If(c.title != nil, HStack(c.title, Spacer(), c.titleX).Alignment(Leading).BackgroundColor(ColorCardTop).Frame(Frame{}.FullWidth()).Padding(Padding{Left: L20, Right: L20, Top: L12, Bottom: L12})),
		VStack(
			c.preBody,
			c.body,
			If(c.footer != nil, HLineWithColor(ColorAccent)),
			HStack(c.footer).Alignment(Trailing).Frame(Frame{}.FullWidth()),
		).
			Frame(c.frame). // TODO this vs ...
			Padding(Padding{Left: L20, Top: L16, Right: L20, Bottom: L20}),
	).
		BackgroundColor(ColorCardBody).
		Border(Border{}.Radius(L20).Elevate(4)).
		Frame(Frame{MinWidth: c.frame.MinWidth, Width: c.frame.Width, MaxWidth: "calc(100% - 2rem)"}) // TODO ... this looks wrong

	if c.disableBoxLayout {
		return stack.Render(ctx)
	}

	dlg := BoxAlign(c.alignment, stack).
		BackgroundColor(colors.M5.WithTransparency(40)).Padding(c.modalPadding)

	return dlg.Render(ctx)
}
