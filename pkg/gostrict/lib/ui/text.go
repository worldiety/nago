// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// TextAlignment aligns the lines of a text.
type TextAlignment uint32

const (
	TextAlignInherit TextAlignment = 0
	TextAlignStart   TextAlignment = 1
	TextAlignEnd     TextAlignment = 2
	TextAlignCenter  TextAlignment = 3
	TextAlignJustify TextAlignment = 4
)

// WordBreak tells, where a word breaks.
type WordBreak int32

const (
	WordBreakInherit  WordBreak = 0
	WordBreakNormal   WordBreak = 1
	WordBreakBreakAll WordBreak = 2
	WordBreakKeepAll  WordBreak = 3
)

// Hyphens tells, whether a word breaks with a hyphen.
type Hyphens int32

const (
	HyphensInherit Hyphens = 0
	HyphensNone    Hyphens = 1
	HyphensManual  Hyphens = 2
	HyphensAuto    Hyphens = 3
)

// WhiteSpace tells, how white space and line breaks are kept.
type WhiteSpace uint32

const (
	WhiteSpaceNormal WhiteSpace = 0
	WhiteSpaceNoWrap WhiteSpace = 1
	// WhiteSpacePreWrap keeps line breaks and indentation, e.g. of source code, and still wraps long lines.
	WhiteSpacePreWrap WhiteSpace = 2
)

// LinkTargetNewWindowOrTab opens a link in a new window or tab.
const LinkTargetNewWindowOrTab = "_blank"

// TText is a text, see Text.
type TText struct{ ref js.Ref }

func (TText) isView() {}

// Text returns a text.
//
//gostrict:host
func Text(content string) TText {
	return TText{ref: wrap(nui.Text(content))}
}

// LinkWithAction returns a text, which looks like a link and calls action, when the user clicks it.
//
//gostrict:host
func LinkWithAction(text string, action func()) TText {
	return TText{ref: wrap(nui.LinkWithAction(text, action))}
}

// Link returns a text, which links to href, e.g. with the target LinkTargetNewWindowOrTab.
//
//gostrict:host
func Link(wnd Window, text string, href string, target string) TText {
	return TText{ref: wrap(nui.Link(wnd.nago(), text, href, target))}
}

// MailTo returns a link, which writes a mail to the email.
//
//gostrict:host
func MailTo(wnd Window, name string, email string) TText {
	return TText{ref: wrap(nui.MailTo(wnd.nago(), name, email))}
}

// Text sets the content.
//
//gostrict:host
func (c TText) Text(content string) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Text(content))}
}

// Underline underlines the text.
//
//gostrict:host
func (c TText) Underline(b bool) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Underline(b))}
}

// Padding sets the padding.
//
//gostrict:host
func (c TText) Padding(padding Padding) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Padding(padding.nago()).(nui.TText))}
}

// WordBreak sets, where a word breaks.
//
//gostrict:host
func (c TText) WordBreak(wordBreak WordBreak) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).WordBreak(nui.WordBreak(wordBreak)))}
}

// Frame sets the size.
//
//gostrict:host
func (c TText) Frame(frame Frame) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Frame(frame.nago()).(nui.TText))}
}

// WithFrame changes the size with fn, which gets the current one.
//
//gostrict:host
func (c TText) WithFrame(fn func(Frame) Frame) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(nui.TText))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TText) FullWidth() TText {
	return TText{ref: wrap(get[nui.TText](c.ref).FullWidth())}
}

// Resolve translates the content, if it is a key of a resource.
//
//gostrict:host
func (c TText) Resolve(b bool) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Resolve(b))}
}

// Hyphens sets the hyphenation.
//
//gostrict:host
func (c TText) Hyphens(h Hyphens) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Hyphens(nui.Hyphens(h)))}
}

// Border sets the border.
//
//gostrict:host
func (c TText) Border(border Border) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Border(border.nago()).(nui.TText))}
}

// LineBreak lets the text wrap.
//
//gostrict:host
func (c TText) LineBreak(lb bool) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).LineBreak(lb))}
}

// HoveredBorder sets the border, while the pointer is over the text.
//
//gostrict:host
func (c TText) HoveredBorder(border Border) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).HoveredBorder(border.nago()))}
}

// PressedBorder sets the border, while the text is pressed.
//
//gostrict:host
func (c TText) PressedBorder(border Border) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).PressedBorder(border.nago()))}
}

// FocusedBorder sets the border, while the text has the focus.
//
//gostrict:host
func (c TText) FocusedBorder(border Border) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).FocusedBorder(border.nago()))}
}

// TextAlignment aligns the lines.
//
//gostrict:host
func (c TText) TextAlignment(align TextAlignment) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).TextAlignment(nui.TextAlignment(align)))}
}

// Visible shows or hides the text.
//
//gostrict:host
func (c TText) Visible(visible bool) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Visible(visible).(nui.TText))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TText) AccessibilityLabel(label string) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).AccessibilityLabel(label).(nui.TText))}
}

// Font sets the font.
//
//gostrict:host
func (c TText) Font(font Font) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Font(font.nago()))}
}

// Color sets the color.
//
//gostrict:host
func (c TText) Color(color Color) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Color(nui.Color(color)))}
}

// BackgroundColor sets the color of the background.
//
//gostrict:host
func (c TText) BackgroundColor(backgroundColor Color) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).BackgroundColor(nui.Color(backgroundColor)).(nui.TText))}
}

// Action calls f, when the user clicks the text.
//
//gostrict:host
func (c TText) Action(f func()) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Action(f))}
}

// LabelFor makes the text the label of the input with the id.
//
//gostrict:host
func (c TText) LabelFor(id string) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).LabelFor(id).(nui.TText))}
}

// Link makes the text a link.
//
//gostrict:host
func (c TText) Link(url, target string) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Link(url, target))}
}

// WhiteSpace tells, how white space and line breaks are kept.
//
//gostrict:host
func (c TText) WhiteSpace(whiteSpace WhiteSpace) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).WhiteSpace(nui.WhiteSpace(whiteSpace)))}
}

// Ellipsis cuts the text with an ellipsis, if it does not fit.
//
//gostrict:host
func (c TText) Ellipsis(ellipsis bool) TText {
	return TText{ref: wrap(get[nui.TText](c.ref).Ellipsis(ellipsis))}
}
