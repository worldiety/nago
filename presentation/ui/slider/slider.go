// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package slider

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// TSlider is a basic component to input number within a given range.
type TSlider struct {
	label          string               // label displayed above or next to the slider
	value          float64              // static number value (used if InputValue is not set)
	inputValue     *core.State[float64] // bound state for controlled input
	supportingText string               // helper text shown by the slider
	errorText      string               // error message shown by the slider
	disabled       bool                 // disables user interaction
	frame          ui.Frame             // layout constraints
	step           float64              // step size to increase/decrease the value
	maxValue       float64              // max value
	minValue       float64              // min value
	showMarkers    bool                 // show markers on the slider, if true
	unit           string               // unit to be displayed next to the value
}

// Slider creates a new slider with the given range.
func Slider(min, max float64) TSlider {
	if min > max {
		min, max = max, min
	}

	c := TSlider{
		value:    min,
		minValue: min,
		maxValue: max,
	}

	return c
}

// Value sets a static value for the slider.
func (c TSlider) Value(value float64) TSlider {
	c.value = value
	return c
}

// SupportingText sets helper text for the slider.
func (c TSlider) SupportingText(text string) TSlider {
	c.supportingText = text
	return c
}

// ErrorText sets an error message for the slider.
func (c TSlider) ErrorText(text string) TSlider {
	c.errorText = text
	return c
}

// Label sets the label text of the slider.
func (c TSlider) Label(label string) TSlider {
	c.label = label
	return c
}

// InputValue binds the slider to a reactive state.
func (c TSlider) InputValue(input *core.State[float64]) TSlider {
	c.inputValue = input
	return c
}

// Disabled disables or enables the slider.
// When disabled, the user cannot interact with it.
func (c TSlider) Disabled(disabled bool) TSlider {
	c.disabled = disabled
	return c
}

// Frame sets the layout frame of the slider (size, width, height, etc.).
func (c TSlider) Frame(frame ui.Frame) TSlider {
	c.frame = frame
	return c
}

// Step defines the step size to increase/decrease number values stepwise
func (c TSlider) Step(step float64) TSlider {
	c.step = step
	return c
}

// Max defines the max value of the slider
func (c TSlider) Max(max float64) TSlider {
	c.maxValue = max
	return c
}

// Min defines the min value of the slider
func (c TSlider) Min(min float64) TSlider {
	c.minValue = min
	return c
}

// ShowMarkers defines whether to show markers on the slider
func (c TSlider) ShowMarkers(showMarkers bool) TSlider {
	c.showMarkers = showMarkers
	return c
}

// Unit defines the unit to be displayed next to the value
func (c TSlider) Unit(unit string) TSlider {
	c.unit = unit
	return c
}

// Render builds and returns the protocol representation of the slider.
func (c TSlider) Render(_ core.RenderContext) core.RenderNode {
	value := proto.SliderValue{
		From: proto.Float(c.value),
	}
	if c.inputValue != nil {
		value = proto.SliderValue{
			From: proto.Float(c.inputValue.Get()),
		}
	}

	return &proto.Slider{
		Label:          proto.Str(c.label),
		SupportingText: proto.Str(c.supportingText),
		ErrorText:      proto.Str(c.errorText),
		Value:          value,
		InputValue:     c.inputValue.Ptr(),
		Disabled:       proto.Bool(c.disabled),
		Frame: proto.Frame{
			MinWidth:  proto.Length(c.frame.MinWidth),
			MaxWidth:  proto.Length(c.frame.MaxWidth),
			MinHeight: proto.Length(c.frame.MinHeight),
			MaxHeight: proto.Length(c.frame.MaxHeight),
			Width:     proto.Length(c.frame.Width),
			Height:    proto.Length(c.frame.Height),
		},
		Step:        proto.Float(c.step),
		Max:         proto.Float(c.maxValue),
		Min:         proto.Float(c.minValue),
		ShowMarkers: proto.Bool(c.showMarkers),
		Unit:        proto.Str(c.unit),
	}
}
