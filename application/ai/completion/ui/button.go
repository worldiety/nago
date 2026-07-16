// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

// Corner selects which screen corner a floating [ChatButton] is anchored to.
type Corner int

const (
	CornerBottomRight Corner = iota
	CornerBottomLeft
	CornerTopRight
	CornerTopLeft
)

// buttonPanelHeight is the height of the scrollable conversation area inside the floating panel.
const buttonPanelHeight = ui.L400

// TChatButton is a floating button that sits in a screen corner and toggles a [Chat] panel. Construct it with
// [ChatButton] and configure it fluently in the nago value style.
type TChatButton struct {
	opts   ChatOptions
	corner Corner
	icon   core.SVG
	label  string
	zIndex int
}

// ChatButton creates a floating chat button from the given [ChatOptions]. Button-specific settings (corner,
// icon, label, z-index) are configured via the fluent methods, not through ChatOptions.
func ChatButton(opts ChatOptions) TChatButton {
	return TChatButton{
		opts:   opts,
		corner: CornerBottomRight,
		icon:   icons.MessageDots,
		label:  "KI-Assistent",
		zIndex: 10,
	}
}

// Corner anchors the button (and its panel) to the given screen corner. Defaults to [CornerBottomRight].
func (b TChatButton) Corner(c Corner) TChatButton {
	b.corner = c
	return b
}

// Icon overrides the button's icon. Defaults to a chat bubble.
func (b TChatButton) Icon(icon core.SVG) TChatButton {
	b.icon = icon
	return b
}

// Label overrides the button's title / accessibility label.
func (b TChatButton) Label(label string) TChatButton {
	b.label = label
	return b
}

// ZIndex overrides the stacking order of the floating container. Defaults to 10.
func (b TChatButton) ZIndex(z int) TChatButton {
	b.zIndex = z
	return b
}

func (b TChatButton) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()
	open := core.AutoState[bool](wnd)

	title := b.opts.Title
	if title == "" {
		title = b.opts.Provider.Name()
	}

	btn := ui.PrimaryButton(func() {
		open.Set(!open.Get())
	}).PreIcon(b.icon).
		AccessibilityLabel(b.label).
		Title(b.label)

	var panel core.View
	if open.Get() {
		panel = chatFrame(
			chatBody(wnd, b.opts, buttonPanelHeight),
			title,
			nil,
			open,
		)
	}

	// Alignment of the button/panel stack depends on which side the corner is on.
	alignment := ui.Trailing
	if b.corner == CornerBottomLeft || b.corner == CornerTopLeft {
		alignment = ui.Leading
	}

	return ui.VStack(
		ui.If(panel != nil, panel),
		btn,
	).Gap(ui.L8).
		Alignment(alignment).
		Position(cornerPosition(b.corner, b.zIndex)).
		Padding(ui.Padding{}.All(ui.L16)).
		Render(ctx)
}

// cornerPosition builds the fixed-position anchor for the given corner.
func cornerPosition(c Corner, zIndex int) ui.Position {
	p := ui.Position{Type: ui.PositionFixed, ZIndex: zIndex}
	switch c {
	case CornerBottomRight:
		p.Right, p.Bottom = ui.L16, ui.L16
	case CornerBottomLeft:
		p.Left, p.Bottom = ui.L16, ui.L16
	case CornerTopRight:
		p.Right, p.Top = ui.L16, ui.L16
	case CornerTopLeft:
		p.Left, p.Top = ui.L16, ui.L16
	}
	return p
}
