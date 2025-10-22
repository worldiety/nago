// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/markdown"
)

type MessageStyle int

const (
	MessageAgent MessageStyle = iota
	MessageHuman
)

type TChatMessage struct {
	style    MessageStyle
	text     string
	markdown bool
	ico      core.SVG
}

func ChatMessage() TChatMessage {
	return TChatMessage{}
}

func (c TChatMessage) Text(text string) TChatMessage {
	c.text = text
	return c
}

func (c TChatMessage) Markdown(text string) TChatMessage {
	c.text = text
	c.markdown = true
	return c
}

func (c TChatMessage) Icon(ico core.SVG) TChatMessage {
	c.ico = ico
	return c
}

func (c TChatMessage) Style(style MessageStyle) TChatMessage {
	c.style = style
	return c
}

func (c TChatMessage) Render(ctx core.RenderContext) core.RenderNode {
	var color ui.Color
	if c.style == MessageHuman {
		color = ui.M3
	}

	return ui.VStack(
		func() core.View {
			if !c.markdown {
				ui.Text(c.text)
			}

			return markdown.Render(markdown.Options{Window: ctx.Window(), RichText: true, TrimParagraph: true}, []byte(c.text))
		}(),

		ui.If(len(c.ico) != 0, ui.ImageIcon(c.ico)),
	).
		BackgroundColor(color).
		Border(ui.Border{}.Width(ui.L1).Color(ui.M5).Radius(ui.L24)).
		Padding(ui.Padding{}.All(ui.L12)).
		Render(ctx)
}
