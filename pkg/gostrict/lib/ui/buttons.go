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
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
)

// ButtonStyle is the style of a button.
type ButtonStyle uint32

const (
	ButtonStylePrimary           ButtonStyle = 1
	ButtonStyleSecondary         ButtonStyle = 2
	ButtonStyleTertiary          ButtonStyle = 3
	ButtonStylePrimaryCritical   ButtonStyle = 4
	ButtonStyleSecondaryCritical ButtonStyle = 5
	ButtonStyleTertiaryCritical  ButtonStyle = 6
)

// TButton is a button, see PrimaryButton.
type TButton struct{ ref js.Ref }

func (TButton) isView() {}

// Button returns a button of the style, which calls action, when the user clicks it.
//
//gostrict:host
func Button(style ButtonStyle, action func()) TButton {
	return TButton{ref: wrap(nui.Button(nui.ButtonStyle(style), action))}
}

// PrimaryButton returns the button of the main action.
//
//gostrict:host
func PrimaryButton(action func()) TButton {
	return TButton{ref: wrap(nui.PrimaryButton(action))}
}

// SecondaryButton returns the button of a secondary action.
//
//gostrict:host
func SecondaryButton(action func()) TButton {
	return TButton{ref: wrap(nui.SecondaryButton(action))}
}

// TertiaryButton returns the button of a minor action, which looks like a text.
//
//gostrict:host
func TertiaryButton(action func()) TButton {
	return TButton{ref: wrap(nui.TertiaryButton(action))}
}

// Critical styles the button for a destructive action, e.g. deleting.
//
//gostrict:host
func (c TButton) Critical(critical bool) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).Critical(critical))}
}

// HRef makes the button a link.
//
//gostrict:host
func (c TButton) HRef(url URI) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).HRef(core.URI(url)))}
}

// Target sets the target of the link, e.g. LinkTargetNewWindowOrTab.
//
//gostrict:host
func (c TButton) Target(target string) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).Target(target))}
}

// Enabled enables or disables the button.
//
//gostrict:host
func (c TButton) Enabled(b bool) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).Enabled(b))}
}

// Title sets the title.
//
//gostrict:host
func (c TButton) Title(text string) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).Title(text))}
}

// Visible shows or hides the button.
//
//gostrict:host
func (c TButton) Visible(b bool) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).Visible(b))}
}

// Disabled disables or enables the button.
//
//gostrict:host
func (c TButton) Disabled(b bool) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).Disabled(b))}
}

// Font sets the font.
//
//gostrict:host
func (c TButton) Font(font Font) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).Font(font.nago()))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TButton) AccessibilityLabel(label string) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).AccessibilityLabel(label))}
}

// PreIcon shows the icon before the title.
//
//gostrict:host
func (c TButton) PreIcon(svg SVG) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).PreIcon(core.SVG(svg)))}
}

// PostIcon shows the icon after the title.
//
//gostrict:host
func (c TButton) PostIcon(svg SVG) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).PostIcon(core.SVG(svg)))}
}

// Preset sets the style.
//
//gostrict:host
func (c TButton) Preset(preset ButtonStyle) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).Preset(nui.ButtonStyle(preset)))}
}

// Frame sets the size.
//
//gostrict:host
func (c TButton) Frame(frame Frame) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).Frame(frame.nago()))}
}

// FullWidth sets the width to the full width of the parent.
//
//gostrict:host
func (c TButton) FullWidth() TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).FullWidth())}
}

// ID sets the id.
//
//gostrict:host
func (c TButton) ID(id string) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).ID(id))}
}

// Key sets a key, which triggers the button.
//
//gostrict:host
func (c TButton) Key(name, id string) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).Key(name, id))}
}

// Alignment aligns the content.
//
//gostrict:host
func (c TButton) Alignment(alignment Alignment) TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).Alignment(nui.Alignment(alignment)))}
}

// NoWrap keeps the title on a single line.
//
//gostrict:host
func (c TButton) NoWrap() TButton {
	return TButton{ref: wrap(get[nui.TButton](c.ref).NoWrap())}
}
