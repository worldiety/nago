// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package colorpicker

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/pkg/xcolor"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// Mode selects the color space of the area and the hue bar of a [Picker].
type Mode int

const (
	// ModeHCT uses hue, chroma and tone of the HCT color space, which is also the base of the nago themes.
	ModeHCT Mode = iota
	// ModeHSV uses hue, saturation and value of the HSV color space.
	ModeHSV
)

// TPicker is a basic component (Color Picker).
// It lets the user pick a color interactively within a color area, a hue bar and an alpha bar, or by entering it
// as hex, HCT, RGB, HSV, HSL or CMYK. The area uses the HCT color space by default. The picked color is written
// as hex into the bound state, i.e. #rrggbb for an opaque color and #rrggbbaa otherwise. See also [PickerField]
// for a field which opens the picker in a dialog.
type TPicker struct {
	label          string
	supportingText string
	errorText      string
	value          ui.Color
	inputValue     *core.State[ui.Color]
	disabled       bool
	disableAlpha   bool
	invisible      bool
	mode           Mode
	frame          ui.Frame
	id             string
}

// Picker creates a color picker showing the given value. Bind a state with [TPicker.InputValue] to receive the
// picked color.
func Picker(label string, value ui.Color) TPicker {
	return TPicker{label: label, value: value}
}

// InputValue binds the state, into which the picked color is written. Its value takes precedence over the value
// given to [Picker].
func (c TPicker) InputValue(state *core.State[ui.Color]) TPicker {
	c.inputValue = state
	return c
}

// Mode selects the color space of the area and the hue bar. The default is [ModeHCT].
func (c TPicker) Mode(mode Mode) TPicker {
	c.mode = mode
	return c
}

// DisableAlpha hides the alpha bar, so that only opaque colors can be picked.
func (c TPicker) DisableAlpha(disable bool) TPicker {
	c.disableAlpha = disable
	return c
}

func (c TPicker) SupportingText(text string) TPicker {
	c.supportingText = text
	return c
}

func (c TPicker) ErrorText(text string) TPicker {
	c.errorText = text
	return c
}

func (c TPicker) Disabled(disabled bool) TPicker {
	c.disabled = disabled
	return c
}

func (c TPicker) Visible(visible bool) TPicker {
	c.invisible = !visible
	return c
}

func (c TPicker) Frame(frame ui.Frame) TPicker {
	c.frame = frame
	return c
}

func (c TPicker) FullWidth() TPicker {
	c.frame = c.frame.FullWidth()
	return c
}

// ID sets an optional identifier, which must be unique within the view tree.
func (c TPicker) ID(id string) TPicker {
	c.id = id
	return c
}

func (c TPicker) Render(ctx core.RenderContext) core.RenderNode {
	value := c.value
	var ptr proto.Ptr
	if c.inputValue != nil {
		value = c.inputValue.Get()

		// the frontend writes into its own state, so that a value which is not a valid hex color never reaches the
		// state of the application
		wire := core.DerivedState[string](c.inputValue, ".colorpicker").Observe(func(newValue string) {
			if color, ok := normalizeHex(newValue, !c.disableAlpha); ok {
				c.inputValue.Update(color)
			}
		})
		ptr = wire.Ptr()
	}

	if normalized, ok := normalizeHex(string(value), !c.disableAlpha); ok {
		value = normalized
	}

	return &proto.ColorPicker{
		Label:          proto.Str(c.label),
		SupportingText: proto.Str(c.supportingText),
		ErrorText:      proto.Str(c.errorText),
		Value:          proto.Str(value),
		InputValue:     ptr,
		Disabled:       proto.Bool(c.disabled),
		Frame:          frameOf(c.frame),
		Mode:           proto.ColorPickerMode(c.mode),
		DisableAlpha:   proto.Bool(c.disableAlpha),
		Invisible:      proto.Bool(c.invisible),
		Id:             proto.Str(c.id),
	}
}

// normalizeHex returns the color as lower case hex, i.e. #rrggbb for an opaque color and #rrggbbaa otherwise. If
// alpha is false, the alpha channel is dropped. It reports false, if the color is not a hex color.
func normalizeHex(v string, alpha bool) (ui.Color, bool) {
	v = strings.TrimSpace(v)
	if len(v) != 7 && len(v) != 9 {
		return "", false
	}

	for _, r := range v[1:] {
		if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
			return "", false
		}
	}

	c, err := xcolor.ParseHex(v)
	if err != nil {
		return "", false
	}

	if !alpha || c[3] == 0xff {
		return ui.Color(fmt.Sprintf("#%02x%02x%02x", c[0], c[1], c[2])), true
	}

	return ui.Color(fmt.Sprintf("#%02x%02x%02x%02x", c[0], c[1], c[2], c[3])), true
}

func frameOf(f ui.Frame) proto.Frame {
	return proto.Frame{
		MinWidth:  proto.Length(f.MinWidth),
		MaxWidth:  proto.Length(f.MaxWidth),
		MinHeight: proto.Length(f.MinHeight),
		MaxHeight: proto.Length(f.MaxHeight),
		Width:     proto.Length(f.Width),
		Height:    proto.Length(f.Height),
		FlexProperties: proto.FlexProperties{
			Grow:          proto.Bool(f.FlexGrow),
			PreventShrink: proto.Bool(f.FlexPreventShrink),
		},
	}
}
