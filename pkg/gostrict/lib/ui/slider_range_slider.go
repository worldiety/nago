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
	"go.wdy.de/nago/presentation/ui/slider"
)

// SliderRangeSliderValue is the value of a range slider, like slider.RangeSliderValue of nago.
type SliderRangeSliderValue struct {
	// From is the start of the range.
	From float64
	// To is the end of the range.
	To float64
}

// TSliderRangeSlider is a range slider, which inputs a range of numbers, see SliderRangeSlider.
type TSliderRangeSlider struct{ ref js.Ref }

func (TSliderRangeSlider) isView() {}

// SliderRangeSlider returns a range slider of the range from min to max, like slider.RangeSlider of nago.
//
//gostrict:host
func SliderRangeSlider(min, max float64) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(slider.RangeSlider(min, max))}
}

// Value sets a static value of the range slider.
//
//gostrict:host
func (c TSliderRangeSlider) Value(value SliderRangeSliderValue) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(get[slider.TRangeSlider](c.ref).Value(value.nago()))}
}

// SupportingText sets the helper text.
//
//gostrict:host
func (c TSliderRangeSlider) SupportingText(text string) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(get[slider.TRangeSlider](c.ref).SupportingText(text))}
}

// ErrorText sets the error message.
//
//gostrict:host
func (c TSliderRangeSlider) ErrorText(text string) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(get[slider.TRangeSlider](c.ref).ErrorText(text))}
}

// Label sets the label.
//
//gostrict:host
func (c TSliderRangeSlider) Label(label string) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(get[slider.TRangeSlider](c.ref).Label(label))}
}

// Disabled disables or enables the range slider.
//
//gostrict:host
func (c TSliderRangeSlider) Disabled(disabled bool) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(get[slider.TRangeSlider](c.ref).Disabled(disabled))}
}

// Frame sets the layout frame of the range slider.
//
//gostrict:host
func (c TSliderRangeSlider) Frame(frame Frame) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(get[slider.TRangeSlider](c.ref).Frame(frame.nago()))}
}

// Step sets the step size.
//
//gostrict:host
func (c TSliderRangeSlider) Step(step float64) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(get[slider.TRangeSlider](c.ref).Step(step))}
}

// Max sets the max value.
//
//gostrict:host
func (c TSliderRangeSlider) Max(max float64) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(get[slider.TRangeSlider](c.ref).Max(max))}
}

// Min sets the min value.
//
//gostrict:host
func (c TSliderRangeSlider) Min(min float64) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(get[slider.TRangeSlider](c.ref).Min(min))}
}

// ShowMarkers shows or hides the markers on the range slider.
//
//gostrict:host
func (c TSliderRangeSlider) ShowMarkers(showMarkers bool) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(get[slider.TRangeSlider](c.ref).ShowMarkers(showMarkers))}
}

// Unit sets the unit, which is shown next to the value.
//
//gostrict:host
func (c TSliderRangeSlider) Unit(unit string) TSliderRangeSlider {
	return TSliderRangeSlider{ref: wrap(get[slider.TRangeSlider](c.ref).Unit(unit))}
}
