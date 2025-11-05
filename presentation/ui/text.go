// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"log/slog"
	"net/url"
	"strings"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

const LinkTargetNewWindowOrTab = "_blank"

type TextAlignment uint

type Hyphens byte

const (
	HyphensInherit Hyphens = iota
	HyphensNone
	HyphensManual
	HyphensAuto
)

const (
	TextAlignInherit TextAlignment = TextAlignment(proto.TextAlignInherit)
	TextAlignStart   TextAlignment = TextAlignment(proto.TextAlignStart)
	TextAlignEnd     TextAlignment = TextAlignment(proto.TextAlignEnd)
	TextAlignCenter  TextAlignment = TextAlignment(proto.TextAlignCenter)
	TextAlignJustify TextAlignment = TextAlignment(proto.TextAlignJustify)
)

// TText is a basic component (Text).
// This component displays text with customizable styling and interaction options.
// It supports colors, background states, padding, borders, accessibility labels,
// text alignment, and interaction callbacks.
// It can be used for labels, inline text, or as an interactive element (e.g. links).
type TText struct {
	content                string      // the text content
	color                  proto.Color // text color
	backgroundColor        proto.Color // background color
	hoveredBackgroundColor proto.Color // background color when hovered
	pressedBackgroundColor proto.Color // background color when pressed
	focusedBackgroundColor proto.Color // background color when focused
	font                   proto.Font  // font styling (size, weight, etc.)
	invisible              bool        // whether the text is hidden
	onClick                func()      // optional legacy click handler

	padding            proto.Padding // padding around the text
	frame              Frame         // layout frame (size, width, height, etc.)
	border             proto.Border  // border styling
	hoveredBorder      proto.Border  // border styling on hover
	focusedBorder      proto.Border  // border styling when focused
	pressedBorder      proto.Border  // border styling when pressed
	accessibilityLabel string        // accessibility label for screen readers
	textAlignment      TextAlignment // text alignment (left, right, center, justify)
	action             func()        // click/tap action
	lineBreak          bool          // whether text should wrap
	underline          bool          // underline the text
	resolve            bool
	hyphens            Hyphens
}

// MailTo creates a mailto: link text component.
// When clicked, it opens the user's email client with the given email address.
func MailTo(wnd core.Window, name string, email string) TText {
	return Link(wnd, name, "mailto:"+email, "_parent")
}

// LinkWithAction creates an interactive link-like text component.
// It applies underline styling, interactive color, and attaches an action callback.
func LinkWithAction(text string, action func()) TText {
	return Text(text).Underline(true).Action(action).Color(ColorInteractive)
}

// Link performs a best guess based on the given href. If the href starts with http or https
// the window will perform an Open call. Otherwise, a local forward navigation is applied.
func Link(wnd core.Window, text string, href string, target string) TText {
	return LinkWithAction(text, func() {
		if wnd == nil {
			slog.Error("cannot execute link action: window is nil")
			return
		}

		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") || strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
			core.HTTPOpen(wnd.Navigation(), core.URI(href), target)
		} else {
			u, err := url.Parse(href)
			if err != nil {
				slog.Error("Failed to parse href link URL", "err", err.Error(), "href", href)
				wnd.Navigation().ForwardToTarget(core.NavigationPath(href), target, nil)
				return
			}

			q := u.Query()
			if len(q) > 0 {
				tmp := core.Values{}
				for k, v := range q {
					if len(v) > 0 {
						tmp[k] = v[0]
					}
				}
				wnd.Navigation().ForwardToTarget(core.NavigationPath(u.Path), target, tmp)
			} else {
				wnd.Navigation().ForwardToTarget(core.NavigationPath(u.Path), target, nil)
			}

		}
	})
}

func Text(content string) TText {
	return TText{content: content}
}

// Underline underlines the Text.
func (c TText) Underline(b bool) TText {
	c.underline = b
	return c
}

// Padding sets a top, right, bottom and left spacing.
func (c TText) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

// Frame sets the width, minWidth, maxWidth, height, minHeight and maxHeight.
func (c TText) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

// WithFrame sets width, minWidth, maxWidth, height, minHeight and maxHeight using a function.
func (c TText) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

// FullWidth sets the width to 100%.
func (c TText) FullWidth() TText {
	c.frame = Frame{}.FullWidth()
	return c
}

// Resolve tries to resolve the current text content against the window bundle at render time to translate
// its contents. This may cause a lot of redundant or wrong lookups and therefore it is disabled by default.
func (c TText) Resolve(b bool) TText {
	c.resolve = b
	return c
}

func (c TText) Hyphens(h Hyphens) TText {
	c.hyphens = h
	return c
}

// Border draws a Border around the component. It's used to set the Border width, color and radius.
// Fore more information also have a look at the Border component.
func (c TText) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

// LineBreak de-/activates line breaking in between the Text.
func (c TText) LineBreak(lb bool) TText {
	c.lineBreak = lb
	return c
}

// HoveredBorder sets the Border width, color and radius when component is hovered.
func (c TText) HoveredBorder(border Border) TText {
	c.hoveredBorder = border.ora()
	return c
}

// PressedBorder sets the Border width, color and radius when the component is clicked.
func (c TText) PressedBorder(border Border) TText {
	c.pressedBorder = border.ora()
	return c
}

// FocusedBorder sets the Border width, color and radius when the component is focused.
func (c TText) FocusedBorder(border Border) TText {
	c.focusedBorder = border.ora()
	return c
}

// TextAlignment sets the position of the Text. For more information also have a look at TextAlignment.
func (c TText) TextAlignment(align TextAlignment) TText {
	c.textAlignment = align
	return c
}

// Visible decides whether a text is shown.
func (c TText) Visible(visible bool) DecoredView {
	c.invisible = !visible
	return c
}

// AccessibilityLabel sets the label of the text. The content of the label is also displayed in the tooltip that appears when you hover over the Text.
func (c TText) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

// Font sets the size, style and width of the Text.
// For more information also have a look at Font.
func (c TText) Font(font Font) TText {
	c.font = font.ora()
	return c
}

// Color sets the Color of the font.
func (c TText) Color(color Color) TText {
	c.color = color.ora()
	return c
}

// BackgroundColor sets the color of the background.
func (c TText) BackgroundColor(backgroundColor Color) DecoredView {
	c.backgroundColor = backgroundColor.ora()
	return c
}

// Action executes the function when the component is clicked.
func (c TText) Action(f func()) TText {
	c.action = f
	return c
}

func (c TText) Render(ctx core.RenderContext) core.RenderNode {

	value := c.content
	if c.resolve {
		value = ctx.Window().Bundle().Resolve(value)
	}

	var hyphens string
	switch c.hyphens {
	case HyphensAuto:
		hyphens = "auto"
	case HyphensManual:
		hyphens = "manual"
	case HyphensNone:
		hyphens = "none"
	}

	return &proto.TextView{
		Value:              proto.Str(value),
		Color:              c.color,
		BackgroundColor:    c.backgroundColor,
		Font:               c.font,
		OnClick:            ctx.MountCallback(c.onClick),
		Invisible:          proto.Bool(c.invisible),
		Border:             c.border,
		Padding:            c.padding,
		Frame:              c.frame.ora(),
		AccessibilityLabel: proto.Str(c.accessibilityLabel),

		HoveredBackgroundColor: c.hoveredBackgroundColor,
		PressedBackgroundColor: c.pressedBackgroundColor,
		FocusedBackgroundColor: c.focusedBackgroundColor,
		HoveredBorder:          c.hoveredBorder,
		FocusedBorder:          c.focusedBorder,
		PressedBorder:          c.pressedBorder,
		TextAlignment:          proto.TextAlignment(c.textAlignment),
		Action:                 ctx.MountCallback(c.action),
		LineBreak:              proto.Bool(c.lineBreak),
		Underline:              proto.Bool(c.underline),
		Hyphens:                proto.Str(hyphens),
	}
}
