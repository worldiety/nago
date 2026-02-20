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

type alignedComponent struct {
	Component core.View
	Alignment proto.Alignment
}

// BoxLayout defines a layout configuration.
// It specifies which child views are placed at the top, bottom,
// center, leading, trailing, or any of the four corners.
type BoxLayout struct {
	Top            core.View
	Center         core.View
	Bottom         core.View
	Leading        core.View
	Trailing       core.View
	TopLeading     core.View
	TopTrailing    core.View
	BottomLeading  core.View
	BottomTrailing core.View
}

// TBox is a layout component (Box).
// It lays out its children according to BoxLayout rules. By definition,
// the container clips its children. This makes it suitable for overlapping
// layouts, but usually requires absolute height and width. Shadows may require
// extra padding since clipped children cannot extend beyond the container.
type TBox struct {
	children                    []alignedComponent
	backgroundColor             proto.Color
	frame                       Frame
	padding                     proto.Padding
	font                        proto.Font
	border                      proto.Border
	accessibilityLabel          string
	invisible                   bool
	disableOutsidePointerEvents bool
}

// BoxAlign creates a new Box with a single child aligned according to
// the given alignment position (e.g., Top, Center, Bottom, etc.).
func BoxAlign(alignment Alignment, child core.View) TBox {
	var bl BoxLayout
	switch alignment {
	case Top:
		bl.Top = child
	case Center:
		bl.Center = child
	case Bottom:
		bl.Bottom = child
	case Leading:
		bl.Leading = child
	case Trailing:
		bl.Trailing = child
	case TopLeading:
		bl.TopLeading = child
	case TopTrailing:
		bl.TopTrailing = child
	case BottomLeading:
		bl.BottomLeading = child
	case BottomTrailing:
		bl.BottomTrailing = child

	default:
		bl.Center = child
	}

	return Box(bl)
}

// Box is a container, in which the given children will be layout to the according BoxLayout
// rules. Note, that per definition the container clips its children. Thus, if working with shadows,
// you need to apply additional padding. Important: this container requires usually absolute height and width
// attributes and cannot work properly using wrap content semantics, because it intentionally allows overlapping.
func Box(layout BoxLayout) TBox {
	c := TBox{
		children: make([]alignedComponent, 0, 9),
	}

	if layout.Center != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Center,
			Alignment: Center.ora(),
		})
	}

	if layout.Top != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Top,
			Alignment: Top.ora(),
		})
	}

	if layout.Bottom != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Bottom,
			Alignment: Bottom.ora(),
		})
	}

	if layout.Leading != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Leading,
			Alignment: Leading.ora(),
		})
	}

	if layout.Trailing != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.Trailing,
			Alignment: Trailing.ora(),
		})
	}

	if layout.TopLeading != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.TopLeading,
			Alignment: TopLeading.ora(),
		})
	}

	if layout.BottomLeading != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.BottomLeading,
			Alignment: BottomLeading.ora(),
		})
	}

	if layout.TopTrailing != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.TopTrailing,
			Alignment: TopTrailing.ora(),
		})
	}

	if layout.BottomTrailing != nil {
		c.children = append(c.children, alignedComponent{
			Component: layout.BottomTrailing,
			Alignment: BottomTrailing.ora(),
		})
	}

	return c
}

// Padding sets the inner spacing around the box's children.
func (c TBox) Padding(p Padding) DecoredView {
	c.padding = p.ora()
	return c
}

// BackgroundColor sets the background color of the box.
func (c TBox) BackgroundColor(backgroundColor Color) DecoredView {
	c.backgroundColor = backgroundColor.ora()
	return c
}

// Frame sets the layout frame of the box, including size and positioning.
func (c TBox) Frame(fr Frame) DecoredView {
	c.frame = fr
	return c
}

// WithFrame applies a transformation function to the box's frame
// and returns the updated component.
func (c TBox) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

// FullWidth sets the box to span the full available width.
func (c TBox) FullWidth() TBox {
	c.frame.Width = "100%"
	return c
}

// Font sets the font style for text content inside the box.
func (c TBox) Font(font Font) DecoredView {
	c.font = font.ora()
	return c
}

// Border sets the border styling of the box.
func (c TBox) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

// Visible controls the visibility of the box; setting false hides it.
func (c TBox) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

// DisableOutsidePointerEvents controls whether pointer events are disabled outside the box's content.
func (c TBox) DisableOutsidePointerEvents(disable bool) TBox {
	c.disableOutsidePointerEvents = disable
	return c
}

// AccessibilityLabel sets a label used by screen readers for accessibility.
func (c TBox) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

// Render builds and returns the protocol representation of the box,
// including its children, frame, background color, padding, and border.
// Each child is rendered according to its alignment within the box.
func (c TBox) Render(ctx core.RenderContext) core.RenderNode {
	var tmp []proto.AlignedComponent
	for _, child := range c.children {
		tmp = append(tmp, proto.AlignedComponent{
			Component: child.Component.Render(ctx),
			Alignment: child.Alignment,
		})
	}

	return &proto.Box{
		Children:                    tmp,
		Frame:                       c.frame.ora(),
		BackgroundColor:             c.backgroundColor,
		Padding:                     c.padding,
		Border:                      c.border,
		DisableOutsidePointerEvents: proto.Bool(c.disableOutsidePointerEvents),
	}
}
