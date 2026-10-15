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
	"go.wdy.de/nago/presentation/ui/picker"
)

// TPickerButton is a button, which opens a picker, see PickerButton.
type TPickerButton struct{ ref js.Ref }

func (TPickerButton) isView() {}

// PickerButton returns a button, which calls action, when the user clicks it, like picker.Button of nago.
//
//gostrict:host
func PickerButton(action func()) TPickerButton {
	return TPickerButton{ref: wrap(picker.Button(action))}
}

// Content sets the content inside the button, e.g. an icon or a text.
//
//gostrict:host
func (c TPickerButton) Content(content View) TPickerButton {
	return TPickerButton{ref: wrap(get[picker.TButton](c.ref).Content(viewOf(content)))}
}

// Dialog sets the dialog, which is inserted into the button, see picker.TButton.Dialog of nago.
//
//gostrict:host
func (c TPickerButton) Dialog(dialog View) TPickerButton {
	return TPickerButton{ref: wrap(get[picker.TButton](c.ref).Dialog(viewOf(dialog)))}
}

// Padding sets the inner spacing.
//
//gostrict:host
func (c TPickerButton) Padding(padding Padding) TPickerButton {
	return TPickerButton{ref: wrap(get[picker.TButton](c.ref).Padding(padding.nago()).(picker.TButton))}
}

// Frame sets the size and the position.
//
//gostrict:host
func (c TPickerButton) Frame(frame Frame) TPickerButton {
	return TPickerButton{ref: wrap(get[picker.TButton](c.ref).Frame(frame.nago()).(picker.TButton))}
}

// WithFrame changes the size and the position with fn, which gets the current one.
//
//gostrict:host
func (c TPickerButton) WithFrame(fn func(Frame) Frame) TPickerButton {
	return TPickerButton{ref: wrap(get[picker.TButton](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(picker.TButton))}
}

// Border sets the border.
//
//gostrict:host
func (c TPickerButton) Border(border Border) TPickerButton {
	return TPickerButton{ref: wrap(get[picker.TButton](c.ref).Border(border.nago()).(picker.TButton))}
}

// Visible shows or hides the button.
//
//gostrict:host
func (c TPickerButton) Visible(visible bool) TPickerButton {
	return TPickerButton{ref: wrap(get[picker.TButton](c.ref).Visible(visible).(picker.TButton))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TPickerButton) AccessibilityLabel(label string) TPickerButton {
	return TPickerButton{ref: wrap(get[picker.TButton](c.ref).AccessibilityLabel(label).(picker.TButton))}
}
