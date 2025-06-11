// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package picker

import (
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
)

type TButton struct {
	dialog             core.View
	content            core.View
	disabled           bool
	errorText          string
	label              string
	frame              ui.Frame
	supportingText     string
	action             func()
	padding            ui.Padding
	border             ui.Border
	invisible          bool
	accessibilityLabel string
}

func Button(action func()) TButton {
	return TButton{action: action}
}

func (c TButton) Content(content core.View) TButton {
	c.content = content
	return c
}

// Dialog is just inserted into the rendered container as well and is not intended for a regular visible view.
// This is pure optional and for sure you can insert the dialog anywhere else and just ignore this.
// However, putting a normal view here, will break the component.
func (c TButton) Dialog(dialog core.View) TButton {
	c.dialog = dialog
	return c
}

func (c TButton) Padding(padding ui.Padding) ui.DecoredView {
	c.padding = padding
	return c
}

func (c TButton) Frame(frame ui.Frame) ui.DecoredView {
	c.frame = frame
	return c
}

func (c TButton) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	c.frame = fn(c.frame)
	return c
}

func (c TButton) Border(border ui.Border) ui.DecoredView {
	c.border = border
	return c
}

func (c TButton) Visible(visible bool) ui.DecoredView {
	c.invisible = !visible
	return c
}

func (c TButton) AccessibilityLabel(label string) ui.DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c TButton) Render(ctx core.RenderContext) core.RenderNode {
	colors := core.Colors[ui.Colors](ctx.Window())
	borderColor := ui.Color("")
	backgroundColor := ui.Color("")
	if c.disabled {
		borderColor = ""
		backgroundColor = colors.Disabled
	} else {
		borderColor = std.Must(colors.I1.WithChromaAndTone(16, 75))
	}

	var fn func()
	if !c.disabled {
		fn = c.action
	}

	inner := ui.HStack(
		c.dialog,
		c.content,
		ui.Spacer(),
		ui.Image().Embed(heroSolid.ChevronDown).Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
	).Action(fn).HoveredBorder(ui.Border{}.Color(borderColor).Width(ui.L1).Radius("0.375rem")).
		Gap(ui.L8).
		BackgroundColor(backgroundColor).
		Frame(ui.Frame{}.FullWidth()).
		Border(ui.Border{}.Color(ui.M8).Width(ui.L1).Radius("0.375rem")).
		Padding(ui.Padding{}.All(ui.L8))

	return ui.VStack(
		ui.IfElse(c.errorText == "",
			ui.Text(c.label).Font(ui.Font{Size: ui.L14}),
			ui.HStack(
				ui.Image().StrokeColor(ui.SE0).Embed(heroSolid.XMark).Frame(ui.Frame{}.Size(ui.L20, ui.L20)),
				ui.Text(c.label).Font(ui.Font{Size: ui.L16}).Color(ui.SE0),
			),
		),
		inner,
		ui.IfElse(c.errorText == "",
			ui.Text(c.supportingText).Font(ui.Font{Size: "0.75rem"}).Color(ui.ST0),
			ui.Text(c.errorText).Font(ui.Font{Size: "0.75rem"}).Color(ui.SE0),
		),
	).Alignment(ui.Leading).
		Gap(ui.L4).
		Visible(!c.invisible).
		Frame(c.frame).
		Border(c.border).
		Padding(c.padding).
		AccessibilityLabel(c.accessibilityLabel).
		Render(ctx)
}
