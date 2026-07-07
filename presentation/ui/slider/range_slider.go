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

// TRangeSlider is a basic component to input number within a given range.
type TRangeSlider struct {
	label          string                        // label displayed above or next to the range slider
	value          RangeSliderValue              // static number value (used if InputValue is not set)
	inputValue     *core.State[RangeSliderValue] // bound state for controlled input
	supportingText string                        // helper text shown by the range slider
	errorText      string                        // error message shown by the range slider
	disabled       bool                          // disables user interaction
	frame          ui.Frame                      // layout constraints
	step           float64                       // step size to increase/decrease the value
	maxValue       float64                       // max value
	minValue       float64                       // min value
	showMarkers    bool                          // show markers on the range slider, if true
	unit           string                        // unit to be displayed next to the value
}

// RangeSliderValue represents the value of a range slider.
type RangeSliderValue struct {
	From float64
	To   float64
}

// RangeSlider creates a new range slider with the given range.
func RangeSlider(min, max float64) TRangeSlider {
	if min > max {
		min, max = max, min
	}

	c := TRangeSlider{
		value: RangeSliderValue{
			From: min,
			To:   max,
		},
		minValue: min,
		maxValue: max,
	}

	return c
}

// Value sets a static value for the range slider.
func (c TRangeSlider) Value(value RangeSliderValue) TRangeSlider {
	c.value = value
	return c
}

// SupportingText sets helper text for the range slider.
func (c TRangeSlider) SupportingText(text string) TRangeSlider {
	c.supportingText = text
	return c
}

// ErrorText sets an error message for the range slider.
func (c TRangeSlider) ErrorText(text string) TRangeSlider {
	c.errorText = text
	return c
}

// Label sets the label text of the range slider.
func (c TRangeSlider) Label(label string) TRangeSlider {
	c.label = label
	return c
}

// InputValue binds the range slider to a reactive state.
func (c TRangeSlider) InputValue(input *core.State[RangeSliderValue]) TRangeSlider {
	c.inputValue = input
	return c
}

// Disabled disables or enables the range slider.
// When disabled, the user cannot interact with it.
func (c TRangeSlider) Disabled(disabled bool) TRangeSlider {
	c.disabled = disabled
	return c
}

// Frame sets the layout frame of the range slider (size, width, height, etc.).
func (c TRangeSlider) Frame(frame ui.Frame) TRangeSlider {
	c.frame = frame
	return c
}

// Step defines the step size to increase/decrease number values stepwise
func (c TRangeSlider) Step(step float64) TRangeSlider {
	c.step = step
	return c
}

// Max defines the max value of the range slider
func (c TRangeSlider) Max(max float64) TRangeSlider {
	c.maxValue = max
	return c
}

// Min defines the min value of the range slider
func (c TRangeSlider) Min(min float64) TRangeSlider {
	c.minValue = min
	return c
}

// ShowMarkers defines whether to show markers on the range slider
func (c TRangeSlider) ShowMarkers(showMarkers bool) TRangeSlider {
	c.showMarkers = showMarkers
	return c
}

// Unit defines the unit to be displayed next to the value
func (c TRangeSlider) Unit(unit string) TRangeSlider {
	c.unit = unit
	return c
}

// Render builds and returns the protocol representation of the range slider.
func (c TRangeSlider) Render(_ core.RenderContext) core.RenderNode {
	value := proto.SliderValue{
		From: proto.Float(c.value.From),
		To:   proto.Float(c.value.To),
	}
	if c.inputValue != nil {
		value = proto.SliderValue{
			From: proto.Float(c.inputValue.Get().From),
			To:   proto.Float(c.inputValue.Get().To),
		}
	}

	if value.From > value.To {
		value.From, value.To = value.To, value.From
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
		RangeMode:   true,
	}
}
