// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type TCard struct {
	title      string
	children   []core.View
	frame      ui.Frame
	background ui.Color
}

func Card(views ...core.View) TCard {
	return TCard{
		frame:      ui.Frame{}.Larger(),
		background: ui.M2,
		children:   views,
	}
}

func (c TCard) Title(title string) TCard {
	c.title = title
	return c
}

func (c TCard) Append(views ...core.View) TCard {
	c.children = append(c.children, views...)
	return c
}

func (c TCard) Frame(fr ui.Frame) TCard {
	c.frame = fr
	return c
}

func (c TCard) BackgroundColor(background ui.Color) TCard {
	c.background = background
	return c
}

func (c TCard) Render(ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		ui.If(c.title != "", cardTitle(c.title)),
	).Append(c.children...).
		BackgroundColor(c.background).
		Alignment(ui.Leading).
		Frame(c.frame).
		Padding(ui.Padding{}.All(ui.L16)).
		Border(ui.Border{}.Radius(ui.L8)).
		Render(ctx)
}

func cardTitle(title string) core.View {
	return ui.VStack(
		ui.Text(title).Font(ui.HeadlineMedium),
		ui.HLine(),
	).Alignment(ui.Leading)
}
