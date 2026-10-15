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

// TSlider is a slider, which inputs a number within a range, see Slider.
type TSlider struct{ ref js.Ref }

func (TSlider) isView() {}

// Slider returns a slider of the range from min to max, like slider.Slider of nago.
//
//gostrict:host
func Slider(min, max float64) TSlider {
	return TSlider{ref: wrap(slider.Slider(min, max))}
}

// Value sets a static value of the slider.
//
//gostrict:host
func (c TSlider) Value(value float64) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).Value(value))}
}

// SupportingText sets the helper text.
//
//gostrict:host
func (c TSlider) SupportingText(text string) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).SupportingText(text))}
}

// ErrorText sets the error message.
//
//gostrict:host
func (c TSlider) ErrorText(text string) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).ErrorText(text))}
}

// Label sets the label.
//
//gostrict:host
func (c TSlider) Label(label string) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).Label(label))}
}

// InputValue binds the slider to a float state.
//
//gostrict:host
func (c TSlider) InputValue(input FloatState) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).InputValue(input.nago()))}
}

// Disabled disables or enables the slider.
//
//gostrict:host
func (c TSlider) Disabled(disabled bool) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).Disabled(disabled))}
}

// Frame sets the layout frame of the slider.
//
//gostrict:host
func (c TSlider) Frame(frame Frame) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).Frame(frame.nago()))}
}

// Step sets the step size.
//
//gostrict:host
func (c TSlider) Step(step float64) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).Step(step))}
}

// Max sets the max value.
//
//gostrict:host
func (c TSlider) Max(max float64) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).Max(max))}
}

// Min sets the min value.
//
//gostrict:host
func (c TSlider) Min(min float64) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).Min(min))}
}

// ShowMarkers shows or hides the markers on the slider.
//
//gostrict:host
func (c TSlider) ShowMarkers(showMarkers bool) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).ShowMarkers(showMarkers))}
}

// Unit sets the unit, which is shown next to the value.
//
//gostrict:host
func (c TSlider) Unit(unit string) TSlider {
	return TSlider{ref: wrap(get[slider.TSlider](c.ref).Unit(unit))}
}
