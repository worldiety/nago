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

// TTextLayout is a layout component (Text Layout).
// It arranges multiple text (and text-like) child views with shared typography,
// spacing, and alignment. Useful for paragraphs, captions, or any block of text
// that needs consistent styling and optional interaction (action callback).
type TTextLayout struct {
	children        []core.View         // child views (typically text nodes)
	alignment       proto.TextAlignment // text alignment for the block
	backgroundColor proto.Color         // optional background fill
	frame           Frame               // layout constraints (size, min/max)
	gap             proto.Length        // spacing between children
	padding         proto.Padding       // inner spacing around children
	border          proto.Border        // border styling
	invisible       bool                // hides the layout when true
	font            proto.Font          // default font applied to children
	// see also https://www.w3.org/WAI/tutorials/images/decision-tree/
	accessibilityLabel string // label for screen readers describing the text block
	action             func() // optional click/tap action for the entire block
}

// Border applies the given border (widths, radii, colors, shadow) to the layout.
func (c TTextLayout) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

// Visible toggles visibility of the layout. Setting false hides it.
func (c TTextLayout) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

// AccessibilityLabel sets the screen-reader label describing the layout's content or purpose.
func (c TTextLayout) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

// TextLayout performs an inline layouting of multiple text elements. The alignment properties of each
// Text are ignored. Any implementation must support an arbitrary amount of text elements with different
// font settings. However, implementations are also open to support images and any other views, as long
// as they can be rendered inline.
func TextLayout(views ...core.View) TTextLayout {
	return TTextLayout{
		children: views,
	}
}

// Action sets a callback function that is executed when the layout is clicked.
func (c TTextLayout) Action(f func()) TTextLayout {
	c.action = f
	return c
}

// Alignment defines the text alignment within the layout.
func (c TTextLayout) Alignment(alignment TextAlignment) TTextLayout {
	c.alignment = proto.TextAlignment(alignment)
	return c
}

// Font sets the font styling for the text in the layout.
func (c TTextLayout) Font(font Font) TTextLayout {
	c.font = font.ora()
	return c
}

// Frame sets the dimensions and position of the layout.
func (c TTextLayout) Frame(f Frame) DecoredView {
	c.frame = f
	return c
}

// WithFrame modifies the current frame using the provided function.
func (c TTextLayout) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

// FullWidth expands the layout to occupy the full available width.
func (c TTextLayout) FullWidth() TTextLayout {
	c.frame.Width = "100%"
	return c
}

// Padding sets the inner spacing of the layout.
func (c TTextLayout) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

// BackgroundColor sets the background color of the layout.
func (c TTextLayout) BackgroundColor(backgroundColor Color) DecoredView {
	c.backgroundColor = backgroundColor.ora()
	return c
}

// Render converts the text layout into a renderable node.
func (c TTextLayout) Render(ctx core.RenderContext) core.RenderNode {

	return &proto.TextLayout{
		Children:           renderComponents(ctx, c.children),
		Frame:              c.frame.ora(),
		TextAlignment:      c.alignment,
		BackgroundColor:    c.backgroundColor,
		Padding:            c.padding,
		AccessibilityLabel: proto.Str(c.accessibilityLabel),
		Invisible:          proto.Bool(c.invisible),
		Font:               c.font,
		Border:             c.border,

		Action: ctx.MountCallback(c.action),
	}
}
