// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package colorpicker

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

// TPickerField is a composite component (Color Picker Field).
// It shows the current color with a label like a text field. A click opens a dialog with a [Picker], and the
// picked color is written into the bound state, once the user applies it. Without a bound state, the field is
// read-only.
type TPickerField struct {
	label          string
	supportingText string
	errorText      string
	value          ui.Color
	inputValue     *core.State[ui.Color]
	disabled       bool
	disableAlpha   bool
	visible        bool
	mode           Mode
	frame          ui.Frame
}

// PickerField creates a field showing the given value. Bind a state with [TPickerField.InputValue] to let the
// user pick a color.
func PickerField(label string, value ui.Color) TPickerField {
	return TPickerField{label: label, value: value, visible: true}
}

// InputValue binds the state, into which the picked color is written. Its value takes precedence over the value
// given to [PickerField].
func (c TPickerField) InputValue(state *core.State[ui.Color]) TPickerField {
	c.inputValue = state
	return c
}

// Mode selects the color space of the picker. The default is [ModeHSV].
func (c TPickerField) Mode(mode Mode) TPickerField {
	c.mode = mode
	return c
}

// DisableAlpha hides the alpha bar of the picker, so that only opaque colors can be picked.
func (c TPickerField) DisableAlpha(disable bool) TPickerField {
	c.disableAlpha = disable
	return c
}

func (c TPickerField) SupportingText(text string) TPickerField {
	c.supportingText = text
	return c
}

func (c TPickerField) ErrorText(text string) TPickerField {
	c.errorText = text
	return c
}

func (c TPickerField) Disabled(disabled bool) TPickerField {
	c.disabled = disabled
	return c
}

func (c TPickerField) Visible(visible bool) TPickerField {
	c.visible = visible
	return c
}

func (c TPickerField) Frame(frame ui.Frame) TPickerField {
	c.frame = frame
	return c
}

func (c TPickerField) FullWidth() TPickerField {
	c.frame = c.frame.FullWidth()
	return c
}

func (c TPickerField) Render(ctx core.RenderContext) core.RenderNode {
	value := c.value
	interactive := c.inputValue != nil && !c.disabled
	if c.inputValue != nil {
		value = c.inputValue.Get()
	}

	colors := core.Colors[ui.Colors](ctx.Window())
	var dialog core.View
	var action func()
	if interactive {
		presented := core.DerivedState[bool](c.inputValue, ".colorpicker.presented")
		pending := core.DerivedState[ui.Color](c.inputValue, ".colorpicker.pending")
		action = func() {
			pending.Set(c.inputValue.Get())
			presented.Set(true)
		}

		dialog = alert.Dialog(
			c.label,
			Picker("", pending.Get()).
				InputValue(pending).
				Mode(c.mode).
				DisableAlpha(c.disableAlpha).
				FullWidth(),
			presented,
			alert.Width(ui.L480),
			alert.Closeable(),
			alert.Cancel(nil),
			alert.Apply(func() (close bool) {
				c.inputValue.Update(pending.Get())
				return true
			}),
		)
	}

	backgroundColor := ui.Color("")
	hoveredBorder := ui.Border{}
	if c.disabled {
		backgroundColor = colors.Disabled
	} else if interactive {
		hoveredBorder = ui.Border{}.Color(option.Must(colors.I1.WithChromaAndTone(16, 75))).Width(ui.L1).Radius("0.375rem")
	}

	hex := string(value)
	if normalized, ok := normalizeHex(hex, !c.disableAlpha); ok {
		hex = string(normalized)
	}

	inner := ui.HStack(
		dialog,
		swatch(value),
		ui.Text(hex).Font(ui.Monospace),
		ui.Spacer(),
		ui.If(interactive, ui.Image().Embed(heroSolid.ChevronRight).Frame(ui.Frame{}.Size(ui.L16, ui.L16))),
	).Action(action).
		HoveredBorder(hoveredBorder).
		FocusedBorder(ui.Border{}.Color(ui.I0).Width(ui.L1)).
		Gap(ui.L8).
		BackgroundColor(backgroundColor).
		Frame(ui.Frame{}.FullWidth()).
		Border(ui.Border{}.Color(ui.M8).Width(ui.L1).Radius("0.375rem")).
		Padding(ui.Padding{}.All(ui.L8))

	return ui.VStack(
		ui.IfElse(c.errorText == "",
			ui.Text(c.label).Font(ui.Font{Size: ui.L16}),
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
		Visible(c.visible).
		Frame(c.frame).
		Render(ctx)
}

// swatch shows the color as a circle or a no-color sign, if the color is empty.
func swatch(c ui.Color) core.View {
	if c == "" {
		return renderNone()
	}

	return ui.VStack().
		BackgroundColor(c).
		Border(ui.Border{}.Circle().Color(ui.ColorInputBorder).Width(ui.L1)).
		Frame(ui.Frame{}.Size(ui.L24, ui.L24))
}
