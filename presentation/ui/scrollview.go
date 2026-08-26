// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"golang.org/x/text/language"
)

var (
	StrScroll = i18n.MustString("scrollView.scroll", i18n.Values{language.German: "Scrollen", language.English: "Scroll"})
)

type ScrollViewAxis int

func (a ScrollViewAxis) ora() proto.ScrollViewAxis {
	return proto.ScrollViewAxis(a)
}

const (
	ScrollViewAxisVertical   ScrollViewAxis = ScrollViewAxis(proto.ScrollViewAxisVertical)
	ScrollViewAxisHorizontal                = ScrollViewAxis(proto.ScrollViewAxisHorizontal)
	ScrollViewAxisBoth                      = ScrollViewAxis(proto.ScrollViewAxisBoth)
)

type ScrollAnimation int

const (
	ScrollAnimationSmooth ScrollAnimation = iota
	ScrollAnimationInstant
)

type ScrollBehavior int

const (
	ScrollBehaviorAlways ScrollBehavior = ScrollBehavior(proto.ScrollBehaviorAlways) // Always scroll when the content grows
	ScrollBehaviorAuto   ScrollBehavior = ScrollBehavior(proto.ScrollBehaviorAuto)   // Automatically scroll when the content grows and is at the scroll end, otherwise ask
	ScrollBehaviorAsk    ScrollBehavior = ScrollBehavior(proto.ScrollBehaviorAsk)    // Ask the user whether to scroll when the content grows
)

// TScrollView is a layout component (Scroll View).
// It provides a scrollable container for a single child view.
// The scroll direction can be vertical (default) or horizontal.
// Supports customization of frame, position, border, background color, and padding.
type TScrollView struct {
	content           core.View      // the scrollable content
	axis              ScrollViewAxis // scroll direction (vertical/horizontal)
	frame             Frame          // layout frame for size and positioning
	position          Position       // content alignment within the scroll view
	border            Border         // optional border around the scroll view
	backgroundColor   Color          // background color
	padding           Padding        // inner padding
	scrollToView      string
	scrollAnimation   ScrollAnimation
	scrollBehavior    ScrollBehavior
	scrollButtonLabel string
}

// A ScrollView can either be horizontal or vertical. By default, it is vertical.
func ScrollView(content core.View) TScrollView {
	return TScrollView{
		content: content,
		axis:    ScrollViewAxisVertical,
	}
}

func (c TScrollView) Content(content core.View) TScrollView {
	c.content = content
	return c
}

// Axis sets the scroll direction (vertical or horizontal).
func (c TScrollView) Axis(axis ScrollViewAxis) TScrollView {
	c.axis = axis
	return c
}

// Frame sets the layout frame for the scroll view.
func (c TScrollView) Frame(frame Frame) TScrollView {
	c.frame = frame
	return c
}

// Position sets the alignment of the content inside the scroll view.
func (c TScrollView) Position(position Position) TScrollView {
	c.position = position
	return c
}

func (c TScrollView) ScrollToView(scrollToView string, animation ScrollAnimation) TScrollView {
	c.scrollToView = scrollToView
	c.scrollAnimation = animation
	return c
}

// Border applies a border around the scroll view.
func (c TScrollView) Border(border Border) TScrollView {
	c.border = border
	return c
}

// Padding sets the inner padding of the scroll view.
func (c TScrollView) Padding(padding Padding) TScrollView {
	c.padding = padding
	return c
}

// BackgroundColor sets the background color of the scroll view.
func (c TScrollView) BackgroundColor(color Color) TScrollView {
	c.backgroundColor = color
	return c
}

// ScrollBehavior defines how the component should behave when the scrollable content grows
func (c TScrollView) ScrollBehavior(behavior ScrollBehavior) TScrollView {
	c.scrollBehavior = behavior
	return c
}

// ScrollButtonLabel sets the label of scroll button when the component asks whether to scroll
func (c TScrollView) ScrollButtonLabel(label string) TScrollView {
	c.scrollButtonLabel = label
	return c
}

// Render builds and returns the protocol representation of the scroll view.
// It includes the scrollable content, axis (vertical/horizontal),
// frame, alignment, border, background color, and padding.
func (c TScrollView) Render(ctx core.RenderContext) core.RenderNode {
	buttonLabel := c.scrollButtonLabel
	if buttonLabel == "" {
		buttonLabel = StrScroll.Get(ctx.Window())
	}

	return &proto.ScrollView{
		Content:           render(ctx, c.content),
		Axis:              c.axis.ora(),
		Frame:             c.frame.ora(),
		Position:          c.position.ora(),
		Border:            c.border.ora(),
		BackgroundColor:   proto.Color(c.backgroundColor),
		Padding:           c.padding.ora(),
		ScrollIntoView:    proto.Str(c.scrollToView),
		ScrollAnimation:   proto.ScrollAnimation(c.scrollAnimation),
		ScrollBehavior:    proto.ScrollBehavior(c.scrollBehavior),
		ScrollButtonLabel: proto.Str(buttonLabel),
	}
}

func (c TScrollView) FullWidth() TScrollView {
	c.frame.Width = Full
	return c
}
